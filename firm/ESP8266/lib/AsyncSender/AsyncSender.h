#pragma once

#include <AsyncHTTPRequest_Generic.hpp>

class AsyncSender{
  public:
    AsyncSender();
    AsyncSender(PGM_P url, PGM_P place, PGM_P writeKey);
    void begin(const __FlashStringHelper * url, const __FlashStringHelper * place, const __FlashStringHelper * writeKey);
    int send(const char *freqMetric);
    void abort();

  private:
    AsyncHTTPRequest _request;
    const __FlashStringHelper * _url;
    const __FlashStringHelper * _place;
    const __FlashStringHelper * _writeKey;
    bool _sending;
    static void _requestCallbackDispatcher(void* optParm, AsyncHTTPRequest* request, int readyState);
	  void MD5Update_P (md5_context_t *ctx, const __FlashStringHelper *data);
	  const char* BytesToHexStr(const uint8_t * const data, char* strbuf, size_t len);
    void _requestCallback(AsyncHTTPRequest* request, int readyState);

};
