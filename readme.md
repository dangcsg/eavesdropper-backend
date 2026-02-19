# About

***Eavesdropper*** is a simple web app to make it super easy to generate transcripts from real world audio like presential meetings, live speeches etc. It is made with a Ionic + Angular frontend and this golang backend.

# Stack

### Frontend
- Written in typescript
- Ionic
- Angular
- Firebase for web hosting

### Backend
- Written in golang
- Gemini API
- Stripe API
- Firebase for Firestore (NoSQL), Auth
- Google Cloud Storage
- Google Cloud Run  

# Arquitechture overview

## Configurations 
- This is where the BackendMode configuration variable lives, it is used to control if the services use real data or mocks and test tables. It can be applied to any service.
- Also contains the deployment mode variable, usefull because sometimes services may be initialized differently if the code is running inside google cloud.
- Contains other configurations which should not change depending on user usage.

## API
- Router - lists the endpoints and matches them with the respective middlewares and handlers
- Middlewares - extract parameters, call validation services, abort request if necessary.
- Handlers - extract parameters, call services, build and return responses
- Error - Error DTO and writter used to standardize the error responses

## Dtos
- requests - are the dtos used for incoming http requests
- responses - are the dtos used to respond to incoming http requests
- resources - are the DB models in use.

## Errs
- List of errors in use. These do not include any messages to display in the client side. Simply used for handling logic here in the backend side.

## Services
- Individual services in use. These call external API's, perform logic to process data and store it in the db

## Data Service
- Contains the two databases in use. Firestore (main db) and cloud storage (to store audio files). 

### Firestore
- Collections - Types the collections in use, their hierarchical relationship and selects the production or test tables depending on the enviroment in use.
- Operations - Where the db operations are coded. This calls the Collections layer and each function may perform one or multiple read/write operations.

### Cloud Storage
- Contains the methods used to interact with the cloud storage to store and download audio files and to get the manifests files.

# Transcribe process

Google Cloud storage is used to store the audio files.
Each audio has a dedicated storage bucket with the path 'recordingSessions/${this.uid}/${this.sessionId}/'.
Each of these buckets has 2 types of files: Audio chunks and a manifest.json which has the metadata associated with the full audio like the number of chunks and the audio file format.

When the user starts recording in the UI, a session is initiated and the chunks are stored as the audio progresses. When the user stops the recording, the session is finalized with the creation and store of the manifest.json in the google cloud storage bucket.

In the backend, when the user stops recording (at this point, all the aduido chunks are in the cloud), a request is done to the backend to generate the transcript, passing the audio session ID.

This is the transcribe handler in the go codebase. It Basically does this:
- Create a temporary local directory
- Read the manifest file in the audio session bucket
- Download (into the temporary directory) each of the audio chunks (.webm) and join them into a single file
- Convert the joined audio file into a .wav file.
- Check if the user has enough credits to get this transcription.
    - Return an error if he does not.
- Pass the .wav file to the gemini API and request a transcription
- Store a db record with the transcript and some metadata like consumed llm tokens and audio seconds.
- Return the transcript to the UI
- Delete the temporary local directory

Notes: 
1) We use two audio formats .webm and .wav because the eventhough the gemini API supports .wav, the libraries used to record in both IOS and Android are easier to use with .webm
2) This process can be improved to a more robust and version. A future version is mentioned in the improvements appendix.


# Stripe

- We use stripe to handle payments. Currently using a 3 tier subscription service.
- In the stripe dashboard, priceID's are configured. Each user subscription will be associated with a priceID, like an instance of that price associated with a customer and with a user and payment/dealines data.
- At the code level, our stripe service implementation exposes methods to 3 destinations: 
    1) Our own API consumed by the frontend, for methods like CreateCustomer.
    2) Backend internal methods, like CheckUserSubscription, to see if he has an active one.
    3) The stripe webhook this golang backend exposes. This is used to listen and react to events coming directly from stripe and keep our backend in sync with actions the users do in the stripe portal like making payments and etc.

### Test in development envioremnt

## Test webhook running locally
We setup a event forwarder to the stripe webhook running in localhost

* stripe login
* stripe listen --forward-to localhost:8080/stripe/webhook
* stripe trigger invoice.paid
    todo - add mocked data to the request. It's not hard

## Test webhook deployed
We set the endpoint in the stripe dashboard and the events are forwarded there.

# Deployment

Google Cloud Run is used for deployment.
    Each deploy can overwrite the running service or create a new one.

    1 - Set production env mode on go code (cnfgs/enviroment.go)

	2 - initialize  Google Cloud CL
		on Google Cloud sdk shell: 
            gcloud config set project PROJECT_ID
            (PROJECT_ID=(hiddenFromThisRepo) , visible in firebase console)

    (open terminal at the code local repository)

	3 - go build

	4 - gcloud run deploy
        choose proj directory
        set service name
        choose appropriate region (eu west 1)
        alloy unauthenticated Invocations - yes

	5 - monitor on https://console.cloud.google.com/run
		    get url from here, pass it to frontend code (IMPORTANT)


# Improvements

### Transcribe process improvement
- Instead of generating the whole transcript at once, passing the full audio to the llm, it can be done iteratively as audio chunks are generated.
- This will bring resilience to mid recording connection problems, usefull for large audios.
- This allows us to check if limits are reached as the user is recording instead of after he finishes recording.
- Probably OpenAI has methods to stream the audio and get the transcript text in a stream, which can then be streamed back to the UI side.



