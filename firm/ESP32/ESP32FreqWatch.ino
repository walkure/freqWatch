#include "driver/mcpwm.h"
#include <float.h>
#include <WiFi.h>
#include <Ambient.h> 
#include <WebServer.h>
#include <ESPmDNS.h>
#include <Wire.h>
#include "AQM1602LCD.h"
#include "const.h"
#include "mbedtls/md.h"
#include <HTTPClient.h>

WiFiClient amb_client;
WebServer server(80);
Ambient ambient;

constexpr uint8_t pivot_freq = 60;

void handleMetrics(void);
bool IRAM_ATTR input_capture_callback(mcpwm_unit_t mcpwm, mcpwm_capture_channel_id_t cap_channel, const cap_event_data_t *edata, void *user_data);
void IRAM_ATTR onInterval();
hw_timer_t* TimerSetup(int id,void (*handler)(), int usec);

void setup() {
  // Setup Serial
  Serial.begin(115200);

  Wire.begin(); // join i2c bus (address optional for master)
  AQMI2CLCD.setup();

  Serial.println("Setting up WiFi");
  // WiFi Setup
  WiFi.begin(wifi_ssid, wifi_passwd);
  while (WiFi.status() != WL_CONNECTED) {  //  Wi-Fiアクセスポイントへの接続待ち
      delay(500);
  }

  MDNS.begin("freqdev");
  Serial.println(WiFi.localIP());
  AQMI2CLCD.setLocate(0,0);
  AQMI2CLCD.print(WiFi.SSID());
  AQMI2CLCD.setLocate(0,1);
  AQMI2CLCD.print(WiFi.localIP());
  // setup ambient
  ambient.begin(ambient_ch, ambient_key, &amb_client);

  // Setup MCPWM to input capture
  Serial.println("Setting up MCPWM init");
  mcpwm_gpio_init(MCPWM_UNIT_0,MCPWM_CAP_0,34);
  mcpwm_capture_config_t input_capture_setup;
  input_capture_setup.cap_edge = MCPWM_NEG_EDGE;
  input_capture_setup.cap_prescale = 1;
  input_capture_setup.capture_cb = input_capture_callback;
  input_capture_setup.user_data = NULL;
  mcpwm_capture_enable_channel(MCPWM_UNIT_0,MCPWM_SELECT_CAP0,&input_capture_setup);
  
  // Webサーバーを起動
  server.on("/metrics", handleMetrics);
  server.onNotFound(handleNotFound);
  server.begin();

  TimerSetup(0,onInterval,1*1000*1000);

  Serial.println("Setting up done");
}

void handleNotFound(void)
{
    server.send(404, "text/plain", "Not Found.");
}

volatile float metricFreq = 0.;
void handleMetrics(void)
{
    String msg;
    char str[23];
    dtostrf(metricFreq, 10, 6, str);
    if(metricFreq < (pivot_freq-10)){
      msg = str;
      msg.concat("Hz is too low\n");
      server.send(500, "text/plain", msg);
      return;
    }else if(metricFreq > (pivot_freq+10)){
      msg = str;
      msg.concat("Hz is too high\n");
      server.send(500, "text/plain", msg);
      return;
    }
    msg = "# HELP power_freq The frequency of power line.\n# TYPE power_freq gauge\npower_freq ";
    msg.concat(str);
    msg.concat("\n");
    
    server.send(200, "text/plain; charset=utf-8", msg);
}

portMUX_TYPE intervalMux = portMUX_INITIALIZER_UNLOCKED;
volatile bool bInterval = false;
void IRAM_ATTR onInterval(){
  portENTER_CRITICAL_ISR(&intervalMux);
  bInterval = true;
  portEXIT_CRITICAL_ISR(&intervalMux);
}

void dumpSerial(float v1,float v2){
  char buff[16];
  dtostrf(v1, 6, 4, buff);
  Serial.print(buff);
  Serial.print(" ");
  dtostrf(v2, 6, 4, buff);
  Serial.println(buff);
}

volatile uint8_t ambientCount = 0;
void get_captured_freqs(float *simple,float *filtered);
void sendFreqMetric(const char *freqMetric);
void loop() {
  server.handleClient();

  if(!bInterval){
    return;
  }
  
  portENTER_CRITICAL_ISR(&intervalMux);
  bInterval = false;
  portEXIT_CRITICAL_ISR(&intervalMux);

  float simple,filtered;
  get_captured_freqs(&simple,&filtered);
  metricFreq = filtered;
  static float old_freq = 0;
  if( fabs(old_freq - filtered) < DBL_EPSILON ) {
    Serial.println("Value not changed");
    AQMI2CLCD.returnHome();
    AQMI2CLCD.print("N.C.           ");
    return;
  }
  //dumpSerial(simple,filtered);
  
  char fbuff[16];
  dtostrf(filtered, 6, 4, fbuff);
  sendFreqMetric(fbuff);
  AQMI2CLCD.returnHome();
  AQMI2CLCD.print(fbuff);
  Serial.print(fbuff);
  
  if(ambientCount > 30){
    ambientCount = 0;
    ambient.set(1,fbuff);
    ambient.send(); 
    AQMI2CLCD.print("Hz *      ");
    Serial.println("Hz sent.");
  }else{
    ambientCount ++;
    AQMI2CLCD.print("Hz        ");
    Serial.println("Hz");
  }
}

portMUX_TYPE caputureCallbackMux = portMUX_INITIALIZER_UNLOCKED;
volatile uint32_t _cap_value = 0; 

volatile uint32_t _cap_sum_value = 0; 
volatile uint16_t _cap_sum_count = 0;

volatile uint32_t _cap_filtered_sum_value = 0;
volatile uint16_t _cap_filtered_sum_count = 0;

constexpr uint32_t filter_floor = APB_CLK_FREQ / (pivot_freq+6); // 1230769
constexpr uint32_t filter_ceil = APB_CLK_FREQ / (pivot_freq-6); // 1777777

bool IRAM_ATTR input_capture_callback(mcpwm_unit_t mcpwm, mcpwm_capture_channel_id_t cap_channel, const cap_event_data_t *edata, void *user_data)
{
  static uint32_t old_cap_value = 0;
  
  portENTER_CRITICAL_ISR(&caputureCallbackMux);
  const uint32_t value =  edata->cap_value;
  _cap_value = value - old_cap_value;
  old_cap_value = value;
  _cap_sum_value += _cap_value;
  _cap_sum_count ++;

  if(_cap_value > filter_floor && _cap_value < filter_ceil){
    _cap_filtered_sum_value += _cap_value;
    _cap_filtered_sum_count ++;
  }
  
  portEXIT_CRITICAL_ISR(&caputureCallbackMux);
  
  return false;
}

void get_captured_freqs(float *simple,float *filtered){
  portENTER_CRITICAL_ISR(&caputureCallbackMux);
  auto sum_value = _cap_sum_value;
  auto sum_count = _cap_sum_count;
  auto filtered_sum_value = _cap_filtered_sum_value;
  auto filtered_sum_count = _cap_filtered_sum_count;

  _cap_sum_count = 0;
  _cap_sum_value = 0;
  _cap_filtered_sum_value = 0;
  _cap_filtered_sum_count = 0;
  portEXIT_CRITICAL_ISR(&caputureCallbackMux);

  auto average_count = (float)sum_value / sum_count;
  *simple = (float)APB_CLK_FREQ / average_count;

  if(filtered_sum_count > 0){
    average_count = (float)filtered_sum_value / filtered_sum_count;
    *filtered = (float)APB_CLK_FREQ / average_count;
  }else{
    *filtered = 0.;
  }
}

const char* BytesToHexStr(const uint8_t * const data, char* strbuf, size_t len);
void sendFreqMetric(const char *freqMetric){
  mbedtls_md_context_t ctx; 
  // https://github.com/wolfeidau/mbedtls/blob/master/mbedtls/md.h#L39
  mbedtls_md_type_t md_type = MBEDTLS_MD_MD5;
  
  uint8_t hashResult[16];
  mbedtls_md_init_ctx(&ctx,mbedtls_md_info_from_type(md_type));
  mbedtls_md_starts(&ctx);
  mbedtls_md_update(&ctx, (const unsigned char *) metric_place, strlen(metric_place));
  mbedtls_md_update(&ctx, (const unsigned char *) freqMetric, strlen(freqMetric));
  mbedtls_md_update(&ctx, (const unsigned char *) metric_key, strlen(metric_key));
  mbedtls_md_finish(&ctx, hashResult);
  mbedtls_md_free(&ctx);
  char hexResult[33];
  BytesToHexStr(hashResult,hexResult,16);

  HTTPClient http;
  auto dataUri = String(metric_base);
  dataUri.concat("?place=");
  dataUri.concat(metric_place);
  dataUri.concat("&freq=");
  dataUri.concat(freqMetric);
  dataUri.concat("&sign=");
  dataUri.concat(hexResult);
  http.begin(dataUri);
  http.setConnectTimeout(100);
  auto httpCode = http.GET();
  if(httpCode < 0){
    Serial.print("CE:");
    Serial.print(httpCode);
    Serial.print(" ");
    return;
  }
  http.end();
  
  if(httpCode == HTTP_CODE_OK){
    return;
  }
  
  Serial.print("HE:");
  Serial.print(httpCode);
  Serial.print(" ");
}

hw_timer_t* TimerSetup(int id,void (*handler)(), int usec)
{
  hw_timer_t* timer;
  timer = timerBegin(id, APB_CLK_FREQ/(1000*1000), true);
  timerAttachInterrupt(timer, handler, true);
  timerAlarmWrite(timer, usec, true);
  timerAlarmEnable(timer);

  return timer;
}

inline char itoc(const uint8_t i) {
    return i < 10 ? i+0x30 : i+0x57; // 0x57 = 0x61('a') - 0x0a
    //return i < 10 ? i+0x30 : i+0x37; // 0x37 = 0x41('A') - 0x0a
}

// bytes列をhex stringに変換
// strbufのサイズは len*2+1
const char* BytesToHexStr(const uint8_t * const data, char* strbuf, size_t len){
    for(size_t i = 0 ; i < len ; i ++){
        strbuf[i*2+1] = itoc(data[i] & 0xf);
        strbuf[i*2] = itoc(data[i] >> 4);
    }
    strbuf[len*2] = '\0';
    return strbuf;
}
