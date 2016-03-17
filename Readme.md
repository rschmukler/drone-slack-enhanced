# drone-slack-enhanced

An enhanced slack notification for Drone and Gogs.

## Configuring

```yaml
notify:
  slack:
    image: rschmukler/drone-slack
    webhook_url: $$SLACK_WEBHOOK
    vcs: yourgogsinstance.com
```
