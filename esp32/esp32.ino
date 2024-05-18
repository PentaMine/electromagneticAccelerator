struct LoggerConf {
  bool isActive;
  unsigned long startTime;
  float duration;
};

const int q[] = { 39, 35, 33, 26, 21, 18 };
const int b[] = { 23, 17, 25, 14, 19, 5 };

bool states[] = { 0, 0, 0, 0, 0, 0 };

LoggerConf logger = { false, 0, 0 };
String loggerQueue;


void setup() {
  // put your setup code here, to run once:
  for (int i = 0; i < 6; i++) {
    pinMode(q[i], INPUT);
    attachInterrupt(q[i], handleInputChangeInterrupt, CHANGE);
  }
  for (int pin : b) {
    pinMode(pin, OUTPUT);
  }

  Serial.begin(115200);
  Serial.println("st");
}

void loop() {
  if (Serial.available() > 0) {
    String incoming = Serial.readStringUntil('\n');

    handleMessage(incoming);
  }

  handleLogger();
}

void handleInputChangeInterrupt() {
  for (int i = 0; i < 6; i++) {
    states[i] = digitalRead(q[i]);
  }


  if (logger.isActive) {
    loggerQueue += "sr" + generateReadMessage() + ',' + micros() + '\n';
  }
}

void handleLogger() {
  if (!logger.isActive) {
    return;
  }

  if (!loggerQueue.isEmpty()) {
    Serial.print(loggerQueue);
    loggerQueue = "";
  }

  if (millis() > logger.startTime + logger.duration * 1000) {
    logger = { false, 0, 0 };
    Serial.println("slf");
  }
}

String generateReadMessage() {
  String message = "";

  for (bool state : states) {
    message = message + (state ? "1" : "0");
  }

  return message;
}

void handleMessage(String text) {
  char func = text[0];
  String message = text.substring(1);

  switch (func) {
    case 'b':
      handleSetBases(message);
      break;
    case 'r':
      handleReadValues(message);
      break;
    case 'l':
      handleLogValues(message);
      break;
    case 'p':
      Serial.println("sPing" + message);
      break;
    default:
      Serial.println("eUnknownFuction");
  }
}

void handleSetBases(String text) {
  if (text.length() != 6) {
    Serial.println("eInvalidMessageLength");
    return;
  }

  bool setStates[6];

  for (int i = 0; i < 6; i++) {
    char state = text.charAt(i);
    switch (state) {
      case '0':
        setStates[i] = false;
        break;
      case '1':
        setStates[i] = true;
        break;
      default:
        Serial.println("eInvalidBaseState");
        return;
    }
  }

  for (int i = 0; i < 6; i++) {
    digitalWrite(b[i], setStates[i]);
  }

  Serial.println("sb");
}
void handleReadValues(String text) {
  if (text != "") {
    Serial.println("eInvalidParameters");
  }

  Serial.println("sr" + generateReadMessage());
}

void handleLogValues(String text) {
  float duration = text.toFloat();

  if (duration <= 0) {
    if (duration == -1) {
      logger = { false, 0, 0 };
      Serial.println("slf");
      return;
    }

    Serial.println("eInvalidDuration");
    return;
  }

  logger = { true, millis(), duration };

  Serial.println("sls");
}
