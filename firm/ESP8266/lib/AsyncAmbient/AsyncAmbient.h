#pragma once

#include <ArduinoJson.h>

#include <AsyncHTTPRequest_Generic.hpp>

class AsyncAmbient{
  public:
    AsyncAmbient();
    AsyncAmbient(uint16_t channelId, PGM_P writeKey);
    void begin(uint16_t channelId, const __FlashStringHelper * writeKey);
    bool set(uint8_t fieldId,float data);
    bool set(uint8_t fieldId,const char* data);
    bool send();
    void abort();

  private:
    AsyncHTTPRequest _request;
    uint16_t _channelId;
    const __FlashStringHelper * _writeKey;
    bool _sending;
    StaticJsonDocument<256> _jsonDoc;
    static void _requestCallbackDispatcher(void* optParm, AsyncHTTPRequest* request, int readyState);
    void _requestCallback(AsyncHTTPRequest* request, int readyState);

};
