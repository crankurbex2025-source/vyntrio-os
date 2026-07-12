# Architecture Overview

## Systemkontext
Vyntrio OS ist eine lokale Server-Appliance mit integriertem Control Plane, Web-UI und systemnahen Runtime-Services.

## Hauptbausteine
- Installer & Base OS
- API & Worker Services
- Storage Runtime
- Container Runtime
- VM Runtime
- Network Runtime
- License & Update Services
- React Dashboard

## Laufzeitmodell
Der Host führt systemnahe Aktionen über gehärtete Adapter aus. Die Control Plane verwaltet State, Jobs, Policies, Audit Logs und UI-Kommunikation.
