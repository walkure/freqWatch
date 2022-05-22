#include <stdint.h>
#include "AsyncAmbient.h"

#include <AsyncHTTPRequest_Generic.hpp>  

AsyncAmbient::AsyncAmbient(){
}

AsyncAmbient::AsyncAmbient(uint16_t channelId, PGM_P writeKey){
  begin(channelId,reinterpret_cast<const __FlashStringHelper *>(writeKey));
}

void AsyncAmbient::begin(uint16_t channelId, const __FlashStringHelper * writeKey){
  _channelId = channelId;
  _writeKey = writeKey;
  _request.onReadyStateChange(_requestCallbackDispatcher,this);
  _request.setDebug(false);
  _sending=false;
}

bool AsyncAmbient::set(uint8_t fieldId,float data){
  return set(fieldId,String(data,5).c_str());
}

bool AsyncAmbient::set(uint8_t fieldId,const char* data){
  if (fieldId < 1 && fieldId > 8)
    return false;
   
  char index[] = "d0";
  index[1] += fieldId;
  // copy data(if `const` char , arduinojson just store pointer)
  _jsonDoc[index] = const_cast<char*>(data);
  
  return true;  
}

bool AsyncAmbient::send(){
    
  if(_sending)
    return false;

  _sending=true;

  xbuf data;

  if (!(_request.readyState() == readyStateUnsent || _request.readyState() == readyStateDone)){
    goto fail;
  }

  char url[43+7]; // 43 = strlen("http://ambidata.io/api/v2/channels/%u/data")
  sprintf_P(url,PSTR("http://ambidata.io/api/v2/channels/%u/data"),_channelId);

  // internally reset to zero after connection closed.
  _request.setTimeout(30);
  if(!_request.open("POST", url)){
    goto fail;
  }

  _jsonDoc["writeKey"] = _writeKey;
  serializeJson(_jsonDoc,data);
  _jsonDoc.clear();
  
  // ambidata.io requires there below headers.
  _request.setReqHeader(F("Content-Type"),F("application/json"));
  _request.setReqHeader(F("Host"),F("ambidata.io"));

  // not requires but...
  _request.setReqHeader(F("Content-Length"),data.available());

  if(_request.send(&data,data.available()))
    return true;

  fail:
  _sending=false;
  return false;
}

/*static*/ void AsyncAmbient::_requestCallbackDispatcher(void* optParm, AsyncHTTPRequest* request, int readyState){
  ((AsyncAmbient*)(optParm))->_requestCallback(request,readyState);
}

void AsyncAmbient::_requestCallback(AsyncHTTPRequest* request, int readyState){
  // do nothing ;)
  
  if (readyState == readyStateDone) 
  {
    if(request->responseHTTPcode() != 200){
      Serial.print(F("\n[Ambient Send Err]:"));
      Serial.print(request->responseHTTPcode());
      Serial.print(F(":"));
      Serial.println(request->responseText());
    }
    _sending=false;
  }
}


void AsyncAmbient::abort(){
  _request.abort();
}
