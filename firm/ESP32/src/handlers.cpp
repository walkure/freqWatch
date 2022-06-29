#include <Arduino.h>
#include <WebServer.h>

#include <constexpr.h>
#include <handlers.h>

WebServer server(80);

void handleNotFound(void)
{
    server.send(404, "text/plain", "Not Found.");
}

void handleClient()
{
    server.handleClient();
}

volatile float metricFreq = 0.;
void handleMetrics(void)
{
    String msg;
    char str[23];
    dtostrf(metricFreq, 10, 6, str);
    if (metricFreq < (pivot_freq - 10))
    {
        msg = str;
        msg.concat("Hz is too low\n");
        server.send(500, "text/plain", msg);
        return;
    }
    else if (metricFreq > (pivot_freq + 10))
    {
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

void handlerSetup()
{
    server.on("/metrics", handleMetrics);
    server.onNotFound(handleNotFound);
    server.begin();
}
