# Who to Trust

## Description

Provides a way to easily update contacts across multiple characters

## Installation

To install and run the project, ensure you have the following prerequisites:

- Go 1.22.3 or newer installed. You can download and install it from [the official Go website](https://golang.org/dl/).
- EVE Online Developer Application credentials. Set up an application and retrieve the `EVE_CLIENT_ID` and `EVE_CLIENT_SECRET` from [EVE Online Developers](https://developers.eveonline.com/applications).
- Your own base64 encoded secret key for the application (if you lose this, you'll need to clear the data directory and all users will need to reauthenticate).  One option is to use the following command to generate a secret key:

```sh
openssl rand -base64 32 
```

Set the `EVE_CLIENT_ID`, `EVE_CLIENT_SECRET`, and `SECRET_KEY` environment variables:

```sh
export EVE_CLIENT_ID=your_client_id
export EVE_CLIENT_SECRET=your_client_secret
```

## Usage

To run the application, use the following command:

```sh
make run
```

After running the command, access the application at [http://localhost:8080](http://localhost:8080).


## Deployment

After updating the makefile to match your Azure configuration, you can push the container to ACR and update your Azure Container Apps deployment with the following command:
To deploy the application to Azure Container Apps, use the following command:

```sh
make full_deploy
```

## Todo
- tbd

## License

This project is licensed under the MIT License. See the LICENSE file in the repository for more information.
