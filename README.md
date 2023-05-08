# netscaleradc-backup
Backup utility for NetScaler ADC (MPX/VPX)



## Configure targets
In order for netscaleradc-backup to run correctly, the user connecting to the NetScaler ADC environments will require the necessary permissions as outlined below:
- Determine the primary node
- List system backups
- Create system backups
- Delete system backups
- Download system files

All backups are created with the following name during the process: **YYYYMMDD_hhmmss**

Depending on your environment type, run the below commands:

- Standalone node: run the commands below
- High-Availability Pair: run the commands on the primary node
- Cluster: run the commands through Cluster IP address

```
add system user $USERNAME $PASSWORD -externalAuth DISABLED -timeout 60 
add system cmdPolicy CMD_CITRIXADCBACKUP ALLOW "(^show\\s+ha\\s+node\\s+0)|(^show\\s+system\\s+backup\\s+\\d{8}_\\d{6})|(^create\\s+system\\s+backup\\s+\\d{8}_\\d{6})|(^rm\\s+system\\s+backup\\s+\\d{8}_\\d{6}\\.tgz)|(^show\\s+system\\s+file\\s+\\d{8}_\\d{6}\\.tgz\\s+-fileLocation\\s+\"/var/ns_sys_backup\")"
bind system user $USERNAME CMD_NETSCALERADCBACKUP 100
```