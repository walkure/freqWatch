#include <Arduino.h>

#include <float.h>
//#include <WiFi.h>
#include <Ambient.h>
#include <ESPmDNS.h>
#include <Wire.h>
#include "AQM1602LCD.h"
#include <WiFiClient.h>

#include <const.h>

#include <handlers.h>
#include <captures.h>
#include <util.h>
#include <constexpr.h>
#include <sender.h>

WiFiClient amb_client;
Ambient ambient;

void IRAM_ATTR onInterval();

void setup()
{
  // Setup Serial
  Serial.begin(115200);

  Wire.begin(); // join i2c bus (address optional for master)
  AQMI2CLCD.setup();

  Serial.println("Setting up WiFi");
  // WiFi Setup
  WiFi.begin(wifi_ssid, wifi_passwd);
  while (WiFi.status() != WL_CONNECTED)
  { //  Wi-Fiアクセスポイントへの接続待ち
    delay(500);
  }

  MDNS.begin("freqdev");
  Serial.println(WiFi.localIP());
  AQMI2CLCD.setLocate(0, 0);
  AQMI2CLCD.print(WiFi.SSID());
  AQMI2CLCD.setLocate(0, 1);
  AQMI2CLCD.print(WiFi.localIP());
  // setup ambient
  ambient.begin(ambient_ch, ambient_key, &amb_client);

  // Setup MCPWM to input capture
  Serial.println("Setting up MCPWM init");
  setupCapture();

  // Webサーバーを起動
  handlerSetup();

  TimerSetup(0, onInterval, 1 * 1000 * 1000);

  Serial.println("Setting up done");
}

portMUX_TYPE intervalMux = portMUX_INITIALIZER_UNLOCKED;
volatile bool bInterval = false;
void IRAM_ATTR onInterval()
{
  portENTER_CRITICAL_ISR(&intervalMux);
  bInterval = true;
  portEXIT_CRITICAL_ISR(&intervalMux);
}

// 連続でこの回数(≒秒)送信に失敗したら再起動する
const uint16_t maxSendFailures = 300;

volatile uint8_t ambientCount = 0;
volatile extern float metricFreq;
void loop()
{
  handleClient();
  if (!bInterval)
  {
    return;
  }

  portENTER_CRITICAL_ISR(&intervalMux);
  bInterval = false;
  portEXIT_CRITICAL_ISR(&intervalMux);

  float simple, filtered;
  get_captured_freqs(&simple, &filtered);
  metricFreq = filtered;
  static float old_freq = 0;
  if (fabs(old_freq - filtered) < DBL_EPSILON)
  {
    Serial.println("Value not changed");
    AQMI2CLCD.returnHome();
    AQMI2CLCD.print("N.C.           ");
    return;
  }
  // dumpSerial(simple,filtered);

  // Wi-Fiが切れていたら再接続を試みる
  if (WiFi.status() != WL_CONNECTED)
  {
    Serial.println("WiFi disconnected. reconnecting");
    WiFi.reconnect();
  }

  char fbuff[16];
  dtostrf(filtered, 6, 4, fbuff);
  static uint16_t sendFailures = 0;
  if (sendFreqMetric(fbuff))
  {
    if (sendFailures > 0)
    {
      // 復旧したら2行目をIPアドレス表示に戻す
      AQMI2CLCD.setLocate(0, 1);
      AQMI2CLCD.print("                ");
      AQMI2CLCD.setLocate(0, 1);
      AQMI2CLCD.print(WiFi.localIP());
    }
    sendFailures = 0;
  }
  else
  {
    sendFailures++;
    if (sendFailures >= maxSendFailures)
    {
      // 送信失敗が続いたら再起動して復旧を試みる
      Serial.println("too many send failures. restarting");
      ESP.restart();
    }
    // 送信失敗中は2行目に失敗回数を表示
    char ebuff[17];
    snprintf(ebuff, sizeof(ebuff), "SEND ERR %-7u", sendFailures);
    AQMI2CLCD.setLocate(0, 1);
    AQMI2CLCD.print(ebuff);
  }
  AQMI2CLCD.returnHome();
  AQMI2CLCD.print(fbuff);
  Serial.print(fbuff);

  if (ambientCount > 30)
  {
    ambientCount = 0;
    ambient.set(1, fbuff);
    ambient.send();
    AQMI2CLCD.print("Hz *      ");
    Serial.println("Hz sent.");
  }
  else
  {
    ambientCount++;
    AQMI2CLCD.print("Hz        ");
    Serial.println("Hz");
  }
}

