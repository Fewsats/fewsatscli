# fewsatscli

The official command line interface for the Fewsats API. 

Fewsats allows you to monetize your files, data and APIs using the L402 protocol. Learn more about the L402 protocol in this [other repository](https://github.com/Fewsats/awesome-L402).

## Usage

```
NAME:
   Fewsats CLI - Interact with the Fewsats Platform.

USAGE:
   Fewsats CLI [global options] command [command options] 

VERSION:
   v0.2.2

COMMANDS:
   account  Interact with your account.
   apikeys  Interact with api keys.
   storage  Interact with storage services.
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## Configuration

The Fewsats CLI tool can be configured by setting parameters in the `~/.fewsatscli` file based on the `sample.env` file. The most important parameter is the `APIKEY`, which is required for most commands.


## Sign up

To sign up for a Fewsats account, run:
```
fewsatscli account signup
```
You will need to provide an email and a password. After signing up, you will need to create an API key using your user/password, unless you have already configured your `~/.fewsatscli` file with an `APIKEY`.


## Create an API Key

To create a new API key, run:
```
fewsatscli apikeys new
```
API keys expire after 28 days by default. You need a valid API key in your `~/.fewsatscli` configuration file to use most of the commands in this CLI.

## Set up your wallet
Currently, the wallet functionality is only for consuming L402 endpoints (paying to get credentials). 

More wallets will be supported in the future, but for now, you can only use ALBY by setting the `ALBY_TOKEN` environment variable in your `~/.fewsatscli` file.

Make sure your Alby token has permissions to pay invoices.


