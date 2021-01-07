# slack-status-cli

Set your Slack status via the command line

## Requirements

Rather than host a web server that you would need to trust, this command runs one on your local machine to retrieve the OAuth code. This page is hosted with an auto generated, self-signed certificate.

This requires you to have `openssl` installed on your machine and, when the page loads for the first time, it will require you to trust the certificate or ignore the warning.

Here's how to do that on [Firefox](https://support.mozilla.org/en-US/kb/error-codes-secure-websites?as=u&utm_source=inproduct#w_self-signed-certificate). On Chrome, you may have to enable `chrome://flags/#allow-insecure-localhost`.

## Example usage

Set auth token (it will store it in `~/.config/slack-status-cli` or your `$XDG_CONFIG_HOME` dir

    slack-status -auth-token <your auth token>

Set status without emoji

    slack-status Walking the dog

Set status with emoji

    slack-status :walking-the-dog: Walking the dog
    slack-status --emoji :walking-the-dog: Walking the dog

Set status with duration (eg. `10m`, `2h`, `7d12h`)

    slack-status 10m :walking-the-dog: Walking the dog
    slack-status :walking-the-dog: Walking the dog for 10m
    slack-status --duration 10m --emoji :walking-the-dog: Walking the dog

Set status with duration and snooze notifications

    slack-status --snooze --duration 12h --emoji :bed: Good night
    slack-status --snooze :bed: Good night for 12h

Set a status that contains a duration

    # Set status to "On a break" for 5 minutes
    slack-status :sleeping: On a break for 5m
    # Set status to "On a break for 5m"  for 5 minutes
    slack-status --duration 5m :sleeping: On a break for 5m
    # Set status to "On a break for 5m"  with no duration
    slack-status :sleeping: "On a break for 5m"

Clear existing status and snooze durations

    slack-status

Snooze notifications without updating your status

    slack-status --duration 15m --snooze

## Future

I plan to do a bit of work to bundle this for easier distribution and maybe support multiple workspaces.
