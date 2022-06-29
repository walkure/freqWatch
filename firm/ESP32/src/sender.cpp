#include <Arduino.h>

#include <mbedtls/md.h>
#include <util.h>
#include <HTTPClient.h>
#include <const.h>

void sendFreqMetric(const char *freqMetric)
{
    mbedtls_md_context_t ctx;
    // https://github.com/wolfeidau/mbedtls/blob/master/mbedtls/md.h#L39
    mbedtls_md_type_t md_type = MBEDTLS_MD_MD5;

    uint8_t hashResult[16];
    mbedtls_md_init_ctx(&ctx, mbedtls_md_info_from_type(md_type));
    mbedtls_md_starts(&ctx);
    mbedtls_md_update(&ctx, (const unsigned char *)metric_place, strlen(metric_place));
    mbedtls_md_update(&ctx, (const unsigned char *)freqMetric, strlen(freqMetric));
    mbedtls_md_update(&ctx, (const unsigned char *)metric_key, strlen(metric_key));
    mbedtls_md_finish(&ctx, hashResult);
    mbedtls_md_free(&ctx);
    char hexResult[33];
    BytesToHexStr(hashResult, hexResult, 16);

    HTTPClient http;
    auto dataUri = String(metric_base);
    dataUri.concat("?place=");
    dataUri.concat(metric_place);
    dataUri.concat("&freq=");
    dataUri.concat(freqMetric);
    dataUri.concat("&sign=");
    dataUri.concat(hexResult);
    if (!http.begin(dataUri))
    {
        Serial.println("cannot begin");
        return;
    }
    http.setConnectTimeout(100);
    auto httpCode = http.GET();
    if (httpCode < 0)
    {
        Serial.print("CE:");
        Serial.print(httpCode);
        Serial.print(" ");
        return;
    }
    http.end();

    if (httpCode == HTTP_CODE_OK)
    {
        return;
    }

    Serial.print("HE:");
    Serial.print(httpCode);
    Serial.print(" ");
}
