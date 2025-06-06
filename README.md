# DigitalOcean App Test

This is my personal repo for testing code on the DigitalOcean App Platform. It's based on the [DigitalOcean Sample getting started guide](https://github.com/digitalocean/sample-dockerfile). I can't promise any of the code in here works or is safe. 

## Run Locally

To test the application locally:

```sh
docker build -t digitalocean-app-test .
docker run -p 8080:8080 digitalocean-app-test
```

Use `curl` or a browser to navigate to your application with any query modifiers (e.g. `http://localhost:8080/add?a=3.5&b=2.1'`) or files (e.g. ` curl -X POST -F "file=@test.txt" localhost:8080/upload`). You should see the response from your app.

## Initialize App on DigitalOcean (First Deploy)

> **IMPORTANT!** You will need to fork this repository into your own GitHub account; DigitalOcean will want to connect to your GitHub account in order to clone the repo on the backend.

To start, I recommend creating a `production` branch. You can use the `main` branch to deploy your app, but I prefer to use `main` for active development (I've lost count the number of times I've messed up the `main` branch in a project...).

```sh
git checkout -b production
git push --set-upstream origin production
git checkout main
```

 1. Sign in to [cloud.digitalocean.com](https://cloud.digitalocean.com/)
 2. Click **App Platform** under *Manage*
 3. Select **GitHub** in the *Get started* box and click **Create app**
 4. Click **Connect GitHub account**
 5. Follow the on-screen instructions to allow DigitOcean access to your GitHub account. I *highly* recommend only giving it access to selected repositories (e.g. this one).
 6. Back in the App setup page, select this repository (e.g. `ShawnHymel/digitalocean-app-test`)
 7. Select your desired branch (e.g. `production`)
 8. Leave *Autodeploy* selected if you want to simply push code to your `production` branch and have the App be automatically updated
 9. On the next screen, configure your server. I recommend deleting one of the *Web Service* instances (so you have only 1 instance) and setting the *Size* to whatever is the cheapest.
 10. Give your app a name (e.g. `addition-app`)
 11. Click **Create app**

You should see "Building..." on the next page. After a few minutes, you should be able to try out your application.

In a browser, go to the link given at the top of the app page along with any query modifiers (e.g. `https://my-app-dsyyq.ondigitalocean.app/add?a=3.5&b=2.1`) or files (e.g. `curl -X POST -F "file=@test.txt" https://my-app-dsyyq.ondigitalocean.app/upload`). You should see the response from your app.

## Deploy Updates

You can deploy updates to your app by push changes to the `production` branch. Make the desired changes and merge them into the `main` branch (like you would for most GitHub projects). Then, merge `main` into `production`:

```sh
git checkout production
git merge main
git push origin production
git checkout main
```

## Delete App from DigitalOcean App Platform

> **IMPORTANT!** You will want to delete your app lest you want to continue paying the subscription fee.

 1. Sign in to [cloud.digitalocean.com](https://cloud.digitalocean.com/)
 2. Click on your **project** (e.g. "First-Project") in the left navigation bar
 3. Click your **app** (e.g. "digitalocean-app-test")
 3. Click **Actions > Destroy App**

## License

I can't promise any of the code in here works or is safe. But if you really want to use it, it is licensed under the [Zero-Clause BSD license](https://opensource.org/license/0bsd).

> Zero-Clause BSD
> 
> Permission to use, copy, modify, and/or distribute this software for
> any purpose with or without fee is hereby granted.
> 
> THE SOFTWARE IS PROVIDED “AS IS” AND THE AUTHOR DISCLAIMS ALL
> WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES
> OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE
> FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY
> DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN
> AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT
> OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
