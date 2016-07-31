# Demisto alert action settings

action.demisto = [0|1]
* Enable demisto action

param.name = <string>
* Configure value for the incident name.

param.type = <string>
* Configure value for the incident type.

param.occured = <string>
* Configure value for the incident type.

param.labels = <string>
* Configure labels for the incident.

param.details = <string>
* Configure value for the incident details.

action.demisto.param.investigate = [1|0]
* Create investigation for the incident
* Defaults to 0 (do not investigate automatically)

action.demisto.param.severity = [Unknown|Low|Medium|High|Critical]
* Incident severity
