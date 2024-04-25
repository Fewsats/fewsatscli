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

More wallets will be supported in the future, but for now, you can only use ALBY by setting the `ALBY_TOKEN` environment variable in your `~/.fewsatscli` file. The ALBY token needs to have `payments:send` permissions to pay the invoices in your behalf.

Make sure your Alby token has permissions to pay invoices.

## Upload a file

To upload a new file, run:
```
❯ fewsatscli storage upload \
   --name "bitcoin.pdf" \
   --description "Satoshi Nakamoto's groundbreaking Bitcoin whitepaper introducing a peer-to-peer electronic cash system" \
   --price 0.05 \
   --file-path bitcoin.pdf
```

This will return the download URL for the file you uploaded.

```
File uploaded successfully.
Download URL:  https://api.fewsats.com/v0/storage/download/f28bd38d-d522-4e9c-b24f-e35ace731d5f
```

## Download a file

To download a file, you can either download the file using the URL or the file ID.

`fewsatscli storage download https://api.fewsats.com/v0/storage/download/f28bd38d-d522-4e9c-b24f-e35ace731d5f`
`
or
`fewsatscli storage download f28bd38d-d522-4e9c-b24f-e35ace731d5f`

In any case, you should see the price of the file and a payment confirmation. Once you confirm, the invoice will be paid with the associated wallet and the file will be downloaded to the current directory.

```
❯ fewsatscli storage download f28bd38d-d522-4e9c-b24f-e35ace731d5f
URL: https://api.fewsats.com/v0/storage/download/f28bd38d-d522-4e9c-b24f-e35ace731d5f
Lightning invoice price: 79 sats
Do you want to continue? (y/N): y
2024/04/25 11:44:08 DEBUG Paid invoice macaroon="AgELZmV3c2F0cy5jb20CQgAAFyX9IVFiO7h3jp9XnhSGAeLxjYX/pzN9deVFi1cWm7qfvLMofw1UpSQ0220bo4CBS4cD/dRs+0wTFffncol/fgACLGZpbGVfaWQ9YjU0ZTJmZjQtMjFhNC00MGFhLTgzMDItOGFiZTkxOWU2NzVlAAIfZXhwaXJlc19hdD0yMDI0LTA1LTI1VDA5OjQ0OjA1WgAABiDABKr4tgJcADmhUXaRzd45uEqEvWYp9yjVY4AptCbkTQ==" invoice=lnbc790n1pnz5f0xpp5zujl6g23vgamsauwnateu9yxq830rrv9l7nnxlt4u4zck4cknwaqdqqcqzzsxqyz5vqsp56hm4ld4229vem59unmrm0r0vl2g8ktfywqj8qn9kr7h78zn3a0js9qyyssqh5vcntq8g30ddfjfjdcuxe020h3w77ygj3y6tz8tj9cteyrwdkdhygvhatx2l33w35uswg9ttjgmy36cuzhfket085s3h65tc9wyhfcq3vkcdl preimage=17ed7e456f217376052487753041f354eb18f85980dd69d904299e3f3ef2519d
File (bitcoin.pdf) downloaded successfully.
```

