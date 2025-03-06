# slack-status-cli

Set your Slack status via the command line

## Requirements

Rather than host a web server that you would need to trust, this command runs one on your local machine to retrieve the OAuth code. This page is hosted with an embedded, self-signed certificate.

When the page loads for the first time, it will require you to trust the certificate or ignore the warning.

Here's how to do that on [Firefox](https://support.mozilla.org/en-US/kb/error-codes-secure-websites?as=u&utm_source=inproduct#w_self-signed-certificate). On Chrome, you may have to enable `chrome://flags/#allow-insecure-localhost`. If you can't get the link to open in Chrome, you can copy the URL into a different browser and try it there. It should work in Firefox or Safari.

## Setup

  1. Install the command from the [Gitea Releases](https://git.iamthefij.com/iamthefij/slack-status-cli/releases) tab

  2. Run the command! On first run, you will be asked to authenticate and install the app in your workspace

## Example usage

Login (it will store it in `~/.config/slack-status-cli` or your `$XDG_CONFIG_HOME` dir

    slack-status -login
    slack-status -login -domain custom-workspace-name
    slack-status -make-default -domain new-default-workspace

Specify domain to use

    # Use the deafult domain
    slack-status :face_with_cowboy_hat: Howdy partner
    # Specify domain for this status
    slack-status -domain my-workspace :face_with_cowboy_hat: Howdy partner
    # Specify domain for this status and make it the new default
    slack-status -domain my-workspace -make-default :face_with_cowboy_hat: Howdy partner

Set status without emoji

    slack-status Walking the dog

Set status with emoji

    slack-status :walking-the-dog: Walking the dog
    slack-status -emoji :walking-the-dog: Walking the dog

Set status with duration (eg. `10m`, `2h`, `7d12h`)

    slack-status 10m :walking-the-dog: Walking the dog
    slack-status :walking-the-dog: Walking the dog for 10m
    slack-status -duration 10m -emoji :walking-the-dog: Walking the dog

Set status with duration and snooze notifications

    slack-status -snooze -duration 12h -emoji :bed: Good night
    slack-status -snooze :bed: Good night for 12h

Set a status that contains a duration

    # Set status to "On a break" for 5 minutes
    slack-status :sleeping: On a break for 5m
    # Set status to "On a break for 5m"  for 5 minutes
    slack-status -duration 5m :sleeping: On a break for 5m
    # Set status to "On a break for 5m"  with no duration
    slack-status :sleeping: "On a break for 5m"

Clear existing status and snooze durations

    slack-status

Snooze notifications without updating your status

    slack-status -duration 15m -snooze

## Development

Building requires you to have `openssl` installed on your machine and to export `CLIENT_ID` and `CLIENT_SECRET` environment variables. The certificate and client info can be overridden at runtime for development purposes by running with exported variables post compilation as well as placing the your openssl pem files in the application's config directory. See `auth.go` for more information.
