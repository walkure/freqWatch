#include <stdint.h>
#include <stdlib.h>
#include "AsyncSender.h"
#include <md5.h>

#include <AsyncHTTPRequest_Generic.hpp>

AsyncSender::AsyncSender()
{
}

AsyncSender::AsyncSender(PGM_P url, PGM_P place, PGM_P writeKey)
{
  begin(reinterpret_cast<const __FlashStringHelper *>(url),
   reinterpret_cast<const __FlashStringHelper *>(place),
    reinterpret_cast<const __FlashStringHelper *>(writeKey));
}

void AsyncSender::begin(const __FlashStringHelper *url, const __FlashStringHelper *place, const __FlashStringHelper *writeKey)
{
  _url = url;
  _writeKey = writeKey;
  _place = place;
  _request.onReadyStateChange(_requestCallbackDispatcher, this);
  //_request.setDebug(true);
  _sending = false;
}

inline char _itoc(const uint8_t i)
{
  return i < 10 ? i + 0x30 : i + 0x57; // 0x57 = 0x61('a') - 0x0a , 0x37 = 0x41('A') - 0x0a
}

// bytes列をhex stringに変換
// strbufのサイズは len*2+1
const char *AsyncSender::BytesToHexStr(const uint8_t *const data, char *strbuf, size_t len)
{
  for (size_t i = 0; i < len; i++)
  {
    strbuf[i * 2 + 1] = _itoc(data[i] & 0xf);
    strbuf[i * 2] = _itoc(data[i] >> 4);
  }
  strbuf[len * 2] = '\0';
  return strbuf;
}

int AsyncSender::send(const char *freqMetric)
{
  if(_sending){
    return -1;
  }

  _sending = true;

  if (!(_request.readyState() == readyStateUnsent || _request.readyState() == readyStateDone))
  {
    _sending = false;
    return -2;
  }

  md5_context_t ctx;
  MD5Init(&ctx);
  MD5Update_P(&ctx, _place);
  MD5Update(&ctx, (const unsigned char *)freqMetric, strlen(freqMetric));
  MD5Update_P(&ctx, _writeKey);
  uint8_t hashResult[16];
  MD5Final(hashResult, &ctx);
  char hexResult[33];
  BytesToHexStr(hashResult, hexResult, 16);

  auto dataUri = String(_url);
  dataUri.concat(F("?place="));
  dataUri.concat(_place);
  dataUri.concat(F("&freq="));
  dataUri.concat(freqMetric);
  dataUri.concat(F("&sign="));
  dataUri.concat(hexResult);

  // internally reset to zero after connection closed.
  _request.setTimeout(1);
  if (!_request.open("GET", dataUri.c_str()))
  {
    _sending = false;
    return -3;
  }

  if(_request.send())
    return 0;

  _sending = false;
  return -4;
}

#define MD5_LOAD_BLOCK_SIZE 16

void AsyncSender::MD5Update_P(md5_context_t *ctx, const __FlashStringHelper *str)
{
  auto data = reinterpret_cast<PGM_P>(str);
  size_t len = strlen_P(data);
  size_t blockLen = (len / MD5_LOAD_BLOCK_SIZE) * MD5_LOAD_BLOCK_SIZE;
  uint8_t buf[MD5_LOAD_BLOCK_SIZE];

  for (size_t i = 0; i < blockLen; i += MD5_LOAD_BLOCK_SIZE)
  {
    memcpy_P(buf, data + i, MD5_LOAD_BLOCK_SIZE);
    MD5Update(ctx, buf, MD5_LOAD_BLOCK_SIZE);
  }

  if (len != blockLen)
  {
    memcpy_P(buf, data + blockLen, (len - blockLen));
    MD5Update(ctx, buf, (len - blockLen));
  }
}

/*static*/ void AsyncSender::_requestCallbackDispatcher(void *optParm, AsyncHTTPRequest *request, int readyState)
{
  ((AsyncSender *)(optParm))->_requestCallback(request, readyState);
}

void AsyncSender::_requestCallback(AsyncHTTPRequest *request, int readyState)
{
  // do nothing ;)

  if (readyState == readyStateDone)
  {
    if (request->responseHTTPcode() != 200)
    {
      Serial.print(F("\n[Sender Err]:"));
      Serial.print(request->responseHTTPcode());
      Serial.print(F(":"));
      Serial.println(request->responseText());
    }
    _sending = false;
  }else if(readyState == readyStateUnsent){
      Serial.println(F("\n[Sender Err]:state unsent"));
    _sending = false;
  }
}

void AsyncSender::abort(){
  _request.abort();
}
