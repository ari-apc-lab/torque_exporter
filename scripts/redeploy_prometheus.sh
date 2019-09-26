#!/bin/bash

sudo -i kubectl delete deployments. prometheus-deployment
sudo -i kubectl create configmap prometheus-example-cm --from-file=/home/linux/ATOSES_spiros/euxdat-monitor/prometheus/prometheus.yml -o yaml --dry-run | sudo -i kubectl apply -f -
sudo -i kubectl create -f /home/linux/ATOSES_spiros/euxdat-monitor/prometheus/prometheus_just_deploy.yaml

