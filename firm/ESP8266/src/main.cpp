#include <Arduino.h>
#include <Wire.h>
#include <ESP8266WiFi.h>
#include <ESP8266HTTPClient.h>

#include <AsyncAmbient.h>
#include <AsyncSender.h>

// https://github.com/khoih-prog/AsyncHTTPRequest_Generic#howto-fix-multiple-definitions-linker-error
// #define _ASYNC_HTTP_LOGLEVEL_ 5
#include <AsyncHTTPRequest_Generic.h>

#include <ESPAsyncWebServer.h>
#include <Ticker.h>
#include <AQMLCD.h>

//volatile uint32_t _cap_value = 0; 
volatile uint32_t _cap_sum_value = 0; 
volatile uint16_t _cap_sum_count = 0; 
volatile uint32_t _old_cap_value = 0;
volatile uint32_t _cap_noise_count = 0;

volatile bool _loading = false;

constexpr uint32_t cap_max = CPU_CLK_FREQ / (50-10);
constexpr uint32_t cap_min = CPU_CLK_FREQ / (60+10);
constexpr uint8_t I2C_SCL = 14;
constexpr uint8_t I2C_SDA = 5;

IRAM_ATTR void captureIntrHandler(){
  uint32_t ccount;
  asm volatile("rsr %0,ccount":"=a"(ccount));
  auto new_cap_value = ccount - _old_cap_value;
  _old_cap_value = ccount;
  
 if(_loading)
   return;

 if(new_cap_value > cap_max || new_cap_value < cap_min){
   _cap_noise_count++;
    return;
 }
    
  //_cap_value = new_cap_value;
  _cap_sum_value += new_cap_value;
  _cap_sum_count ++;
}

const extern uint16_t ambient_ch;
extern const char endpoint[];
extern const char place[];
extern const char shareKey[];
extern const char writeKey[];

AsyncSender sender(endpoint,place,shareKey);
AsyncAmbient ambient(ambient_ch,writeKey);

constexpr uint8_t FREQ_INPUT_PIN = 12;

float metricsFreq;
void load(){
//  noInterrupts();
  _loading = true;
  //auto cap_value = _cap_value; _cap_value = 0;
  auto cap_sum_value = _cap_sum_value ; _cap_sum_value = 0;
  auto cap_sum_count = _cap_sum_count ; _cap_sum_count = 0;
  _loading = false;
//  interrupts();

  //auto freq  = 80*1000*1000. / cap_value;

  auto afreq = 80*1000*1000. / ((float)cap_sum_value / cap_sum_count);
  metricsFreq = afreq;
  if(!isfinite(afreq)){
    Serial.println("NaN END");
    return;
  }
/* 
  Serial.print(String(freq,4));
  Serial.print(" ");
*/
  auto freqMetric = String(afreq,4);
  //Serial.println(freqMetric);
  auto result = sender.send(freqMetric.c_str());
  if(result < 0){
    Serial.print("send failed.");
    Serial.println(result);
  }
  AQMI2CLCD.returnHome();
  AQMI2CLCD.print(freqMetric);

  static auto ambientCount = 0;
  if(ambientCount > 30){
    ambient.set(2,freqMetric.c_str());
    auto bresult = ambient.send();
    if(bresult){
      //Serial.println("Sent ambient");
    }else{
      Serial.println("Sent ambient. failed...");
    }
    ambientCount = 0;
    AQMI2CLCD.print(F("*"));
  }else{
    AQMI2CLCD.print(F(" "));
    ambientCount ++;
  }

}

void startServer();
extern const char wifiSSID[];
extern const char wifiPass[];

Ticker tickLoad;
bool onLoad = false;
void onLoadHandler() {
  onLoad = true;
}

void onWifiDisconnected(const WiFiEventStationModeDisconnected& event);
void onWifiGotIP(const WiFiEventStationModeGotIP& event);

void setup() {
  Serial.begin(76800);
  Serial.println(F("Waiting for power."));
  Wire.begin(I2C_SDA,I2C_SCL);
  AQMI2CLCD.setup();

  delay(5000);

  Serial.println(F("WiFi setting up"));
  static auto disconnectedHandler = WiFi.onStationModeDisconnected(onWifiDisconnected);
  static auto gotIpEventHandler = WiFi.onStationModeGotIP(onWifiGotIP);
  static auto connectedHandler = WiFi.onStationModeConnected([](const WiFiEventStationModeConnected& event){
    Serial.println(F("WiFi connected..."));
  });
  // WiFi Setup
  WiFi.persistent(false);
  WiFi.begin(FPSTR(wifiSSID), FPSTR(wifiPass));
}


void loop() {
  if(onLoad){
    onLoad = false;
    load();
  }
}

void handleNotFound(AsyncWebServerRequest *request) {
  request->send(404, "text/plain", "Not found");
}

void handleMetrics(AsyncWebServerRequest *request) {
  if(!isfinite(metricsFreq)){
    return;
  }
  auto message = String(F("# HELP power_freq The frequency of power line.\n# TYPE power_freq gauge\npower_freq "));
  message.concat(String(metricsFreq,4));
  message.concat(F("\n"));
  request->send(200, "text/plain", message);
}

AsyncWebServer server(80);

void onWifiDisconnected(const WiFiEventStationModeDisconnected& event){
  Serial.print(F("Disconnected from WIFI access point. reason:"));
  Serial.println(event.reason);

  tickLoad.detach();
  detachInterrupt(FREQ_INPUT_PIN);
  server.end();
  sender.abort();
  ambient.abort();
  Serial.println(F("Reconnecting..."));

  WiFi.begin(FPSTR(wifiSSID), FPSTR(wifiPass));
}

void onWifiGotIP(const WiFiEventStationModeGotIP& event){
  Serial.print(F("Got IP Address:"));
  Serial.println(WiFi.localIP());

  AQMI2CLCD.setLocate(0,0);
  AQMI2CLCD.print(WiFi.SSID());
  AQMI2CLCD.setLocate(0,1);
  auto ip = WiFi.localIP().toString();
  AQMI2CLCD.print(ip.substring(ip.length()-8));

  tickLoad.attach(1, onLoadHandler);

  pinMode(FREQ_INPUT_PIN, INPUT_PULLUP);
  attachInterrupt(FREQ_INPUT_PIN, captureIntrHandler, FALLING );

  server.on("/metrics", HTTP_GET, handleMetrics);
  server.onNotFound(handleNotFound);
  server.begin();
}
