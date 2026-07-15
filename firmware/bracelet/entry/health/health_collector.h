/*
 * Eregen (颐贞) - Health Data Collector Header
 * Collects PPG heart rate + SpO2 and IMU step count, encodes and publishes via MQTT.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef HEALTH_COLLECTOR_H
#define HEALTH_COLLECTOR_H

/**
 * Initialize the health data collection subsystem.
 * Must be called once before health_collect_and_send().
 */
void health_init(void);

/**
 * Collect health data from sensors, encode, and publish via MQTT.
 * Reads PPG (HR + SpO2) and IMU (step count), builds an MSG_HEALTH message,
 * encodes it with message_encode(), and publishes via cat1_mqtt_publish().
 */
void health_collect_and_send(void);

#endif /* HEALTH_COLLECTOR_H */
