# tweet-destroyer

This Cloud Run service trawls through the authenticated user's twitter timeline that it can and deletes the tweets that match a given set of rules.

The current rules are:
* Keep the authenticated user's tweets if they have favourited them;
* Destroy anything more than a week old with no engagement (no RTs, no favourites)
* Destroy anything more than a month old with < 10 RTs/faves
* Destroy anything older than 6 months with < 25 RTs/faves

## There's some manual setup

**To enable access to the Twitter API**: you need a Twitter token, token secret, and an access token and secret for your profile - the latter must have read **and write** access. You can generate them all via the Twitter [developer portal](https://developer.twitter.com/en). I saved all of these credentials as a single JSON blob in Google's [Secret Manager](https://cloud.google.com/secret-manager).

**To use Google Cloud services**: you need to enable several APIs, as needed - specifically, for Cloud Run and Cloud Scheduler.

**To allow the Cloud Run service to read from Secret Manager**: you need to give the service account that is running the Cloud Run container _Secret Manager Secret Accessor_ permissions, to read the secret. You can do this in the Google Cloud Secret Manager console. By default, Cloud Run services or jobs [run as the default Compute Engine service account](https://cloud.google.com/run/docs/configuring/service-accounts).

**To trigger the Cloud Run service with Cloud Scheduler**: you need to give a service account permission to [invoke a cloud run service](https://cloud.google.com/run/docs/triggering/using-scheduler#create-service-account).

## Create & deploy a Cloud Run service

This service was created with the [kettle-cli](https://github.com/operatorai/kettle-cli), which you can install [using brew](https://github.com/nlathia/kettle-cli#installing-with-brew). You can then start from a [template](https://github.com/nlathia/kettle-templates):

```bash
❯ kettle create golang-gcloud-run
Project name: my-project

✅  Created:  <path/to/my-project>
```

I modified the template to use Golang 1.19.1, and added the logic I needed in `tweets.go` (interacting with the Twitter API), `rules.go` (rules to decide whether to keep a tweet), and `handler.go` (to bring it all together)

The service can be built and deployed with:

```bash
❯ cd my-project
❯ kettle deploy .
...

🔍  API Endpoint:  https://<long-url-values>.run.app
✅  Deployed!
```

Finally, you can run this manually with:

```bash
❯ curl -X POST -d '{"dry_run": true, "max_iterations": 2}' https://<long-url-values>.run.app
```

When `dry_run` is set to `true`, no tweets are deleted. Any non-zero value for `max_iterations` limits how many batches of ~200 tweets the service tries to retrieve. With `max_iterations=1` it will only collect one batch of tweets.

## Run the service on a schedule

To run the service to run on a schedule, I set up Cloud Scheduler [manually](https://cloud.google.com/run/docs/triggering/using-scheduler).

Important! By default, the deployment command above creates a resource **that is public and can be accessed by anyone on the Internet** (which is why that `curl` command above works!). You will need to change this: when you deploy the service you are using with Cloud Scheduler, make sure you do NOT allow unauthenticated invocations. 

## How much does this cost?

Cloud Scheduler gives you [3 free jobs per month, per billing account](https://cloud.google.com/scheduler/pricing). A job is not billed for individual executions.

Cloud Run [is free](https://cloud.google.com/run/pricing) up to 180,000 vCPU seconds/month, 360,000 GiB-seconds per month, and 2 million requests.

Cloud Logging [is free](https://cloud.google.com/stackdriver/pricing) up to the first 50 GiB/project ingested per month. Logs retained for the default period don't incur a storage cost.

Elevated access to the Twitter API [is free](https://developer.twitter.com/en/products/twitter-api) and gives 2M tweets/month and 3 app environments.
