#!/bin/bash

sudo -i kubectl delete deployments. torque-exporter
sudo -i kubectl create -f /home/linux/ATOSES_spiros/GITHUB-torque_exporter/torque_exporter/yaml/torque_exporter.yaml

