![](https://badges.fyi/github/license/Luzifer/elb-instance-status)
![](https://badges.fyi/github/latest-release/Luzifer/elb-instance-status)
![](https://badges.fyi/github/downloads/Luzifer/elb-instance-status)

# Luzifer / elb-instance-status

`elb-instance-status` is a small daemon you can run on any instance on an autoscaling group. It periodically executes commands using a `bash` shell and checks for their exit status (0 = fine, everything else = not fine). The collected check results are exposed using an HTTP listener which then can be used by an ELB health check for that machine. This enables your autoscaling-group to react to custom health checks on your machine.

For example given you have a process eating all inodes on your machine and you have no chance to clean up these files you could use this daemon to terminate the instance as soon as the inode usage is too high. Maybe this is a bad example because file system cleanups should be possible all the time but you get the point: Something is wrong on one of your cattle-machines? Remove it.

Checks run every minute by default (`--check-interval`) and each check batch is cancelled shortly before the next interval, so keep checks cheap and faster than the interval; longer or expensive checks should run separately, for example via cron writing a status file read by this daemon.

Check definitions are loaded from `--check-definitions-file` (default: `/etc/elb-instance-status.yml`), which may be a local file or URL and is refreshed every `--config-refresh` (default: `10m`).

If the unhealthy threshold (default: 5 checks) is crossed the HTTP status will switch from 200 (OK) to 500 (Internal Server Error) which will cause the ELB to mark your machine unhealthy and the autoscaling-group will remove that machine. Of course you need to ensure there is a starting grace period to give your machine enough time to settle and get all checks green. And you also need to take care the new machines started as a replacement for the unhealthy ones are going to be healthy. Otherwise your whole cluster gets taken out of service.

## Usage

1. Install the daemon on your machine
2. Write a yaml file containing the checks you want to execute
3. Start the daemon
4. Put an ELB health check on your autoscaling-group using the daemon's `/status` path as the check target

```bash
# curl -is localhost:3000/status
HTTP/1.1 200 OK
Date: Fri, 03 Jun 2016 10:56:13 GMT
Content-Length: 426
Content-Type: text/plain; charset=utf-8

[PASS] Ensure there are at least 30% free inodes on /var/lib/docker
[PASS] Ensure there are at least 30% free inodes on /
[PASS] Ensure docker can start a small container
[PASS] Ensure volume on /var/lib/docker is mounted
[PASS] Ensure there is at least 30% free disk space on /var/lib/docker
```

### Check format

The checks are defined in a quite simple yaml file:

```yaml
root_free_inodes:
  name: Ensure there are at least 30% free inodes on /
  command: test $(df -i | grep "/$" | xargs | cut -d ' ' -f 5 | sed "s/%//") -lt 70
 
lib_docker_mounted:
  name: Ensure volume on /var/lib/docker is mounted
  command: mount | grep -q /var/lib/docker
 
docker_run:
  name: Ensure docker can start a small container
  command: docker run --rm alpine /bin/sh -c "echo testing123" | grep -q testing123
```

They consist of a unique ID and three keys for each check:

- `name` (required), A descriptive name of the check (probably do not use the same name twice)
- `command` (required), The check itself. Needs to have exit code 0 if everything is fine and any other if something is wrong.  
  The checks are executed using `/bin/bash -e -o pipefail -c "<command>"`.
- `warn_only` (optional, default: false), Only put a WARN-line into the output but do not set HTTP status to 500
