#include <Arduino.h>


#include <Wire.h>
#include "SparkFun_ENS160.h"  
#include "SparkFunBME280.h"
#include <WiFi.h>
#include <HTTPClient.h>
#include <WiFiUdp.h>
#include <PubSubClient.h>


SparkFun_ENS160 myENS;
BME280 myBME280;

int ensStatus;

const char* ssid     = "DPSLab";
const char* password = "Asdf1234";
const char* mqtt_server = "192.168.0.86";  // or use your local broker IP


int aq;
int co2;
float tvoc;
float hum;
float temp;

WiFiUDP ntpUDP;

WiFiClient espClient;
PubSubClient client(espClient);

// Reconnect to MQTT broker
void reconnect() {
  while (!client.connected()) {
    Serial.print("Attempting MQTT connection...");
    // You can give your ESP32-S2 a unique client ID
    if (client.connect("ESP32S2Client")) {
      Serial.println("connected");
    } else {
      Serial.print("failed, rc=");
      Serial.print(client.state());
      Serial.println(" trying again in 5 seconds");
      delay(5000);
    }
  }
}

void setup() {
  Wire.begin();

  Serial.begin(115200);

  WiFi.begin(ssid, password);
  Serial.println("Connecting");

  while(WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  
  Serial.println("");
  Serial.print("Connected to WiFi network with IP Address: ");
  Serial.println(WiFi.localIP());
  client.setServer(mqtt_server, 1883);  // Port 1883 for non-TLS

  if (!myENS.begin()) {
    Serial.println("Did not begin.");
    while (1)
      ;
  }

  if (myBME280.beginI2C() == false)  //Begin communication over I2C
  {
    Serial.println("The sensor did not respond. Please check wiring.");
    while (1)
      ;  //Freeze
  }

  // Reset the indoor air quality sensor's settings.
  if (myENS.setOperatingMode(SFE_ENS160_RESET))
    Serial.println("Ready.");

  delay(100);

  // Device needs to be set to idle to apply any settings.
  // myENS.setOperatingMode(SFE_ENS160_IDLE);

  // Set to standard operation
  // Others include SFE_ENS160_DEEP_SLEEP and SFE_ENS160_IDLE
  myENS.setOperatingMode(SFE_ENS160_STANDARD);

  // There are four values here:
  // 0 - Operating ok: Standard Operation
  // 1 - Warm-up: occurs for 3 minutes after power-on.
  // 2 - Initial Start-up: Occurs for the first hour of operation.
  //                                              and only once in sensor's lifetime.
  // 3 - No Valid Output
  ensStatus = myENS.getFlags();
  Serial.print("Gas Sensor Status Flag: ");
  Serial.println(ensStatus);
  delay(5000);
}

void loop() {
  if (!client.connected()) {
    reconnect();
  }
  
  String ipString = WiFi.localIP().toString();

  client.loop();

  if (myENS.checkDataStatus()) {
    Serial.print("Air Quality Index (1-5) : ");
    //Serial.println(myENS.getAQI());
  aq=myENS.getAQI();
  Serial.println(aq);


   Serial.print("CO2 concentration: ");
    //Serial.print(myENS.getECO2());
    //Serial.println("ppm");
    int co2=myENS.getECO2();
    Serial.println(co2);


    Serial.print("Total Volatile Organic Compounds: ");
    tvoc=myENS.getTVOC();
    //Serial.print(myENS.getTVOC());
    Serial.println(tvoc);


   

    Serial.print("Humidity: ");
    hum=myBME280.readFloatHumidity();
    Serial.println(hum);

    //Serial.print("Pressure: ");
   // Serial.print(myBME280.readFloatPressure(), 0);
   // Serial.println("Pa");

    //Serial.print("Alt: ");
    //Serial.print(myBME280.readFloatAltitudeMeters(), 1);
    //Serial.println("meters");
    //Serial.print(myBME280.readFloatAltitudeFeet(), 1);
   // Serial.println("feet");

    Serial.print("Temp: ");
    temp=myBME280.readTempC();
    Serial.println(temp);
    //Serial.print(myBME280.readTempF(), 2);
    //Serial.println(" degF");

    Serial.println();

    char temphum[8];
    dtostrf(hum, 1, 2, temphum); // Convert float to char*
    client.publish("station/1/aq", temphum);

    char tempCo2[6];  // Enough for a 5-digit int and null terminator
    itoa(co2, tempCo2, 10);  // Base 10 conversion
    client.publish("station/1/co2", tempCo2);

    // String serverPath = serverName +"?aq="+String(aq)+"&co2="+String(co2)+"&tvoc="+String(tvoc)+"&hum="+String(hum)+"&temp="+String(temp);

  }
  delay(200);
}