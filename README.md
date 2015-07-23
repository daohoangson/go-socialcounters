# Social Counters with SVG logos!
This golang app provides consistent and faster social counters for popular social networks.
Traditionally, website owners have to include scripts from social networks to display their buttons
and share counters (see [Facebook](https://developers.facebook.com/docs/plugins/share-button),
[Twitter](https://about.twitter.com/resources/buttons) and [Google](https://developers.google.com/+/web/share/) documents).
Unfortunately, those buttons do not align well together (different style, inflexible width, etc.) and more importantly,
they often make a lot of extra requests in order to render correctly, generally make it a bad user experience.

This app implement counters fetcher for Facebook, Twitter and Google+ and return everything in one single request
of about 8KB (gzip) or 15KB (original). With the only script inclusion, it can display all the signature logos of
those networks **and** their counters for the specified url.

## Usage

### all.js (requires jQuery)
After deploying the app to your server, simply include the script in your website to start rendering the counters.
The `all.js` script detect svg support in user browser by itself and switch to use svg logo whenever possible;
for older browsers, data uri will be used.

````
<script src="//socialcounters.domain.com/js/all.js?url=http://www.domain.com"></script>
````

#### Parameters

 * `url` (**required**)
 * `shorten` (_optional_): determine whether the counters should be shortened to use "m" (million) or "k" (thousand) suffices.
 This feature is off by default.
 * `ttl` (_optional_): specify the cache max-age value. Useful if you put your server behind a some caching service or CDN.
 This should help you to control the actual load on the app server.
 * `target` (_optional_): by default, `all.js` looks for `.socialcounters-container` in the DOM and insert its buttons
 into that element. You can use `target` param to change the selector which is used to find the container element.
 When no container is found, `all.js` will append into `body`.

### jQuery Plugin
jQuery Plugin is the advanced option if you want to customize the buttons look and feel to match your website design.
You will need to include the plugin script and layout the elements on the page yourself (see demo for pointers).
Finally, call `socialcounters()` on the jQuery object to get counters data from the server and populate them.

````
<script src="//socialcounters.domain.com/js/jquery.plugin.js"></script>
<link rel="stylesheet" href="//socialcounters.domain.com/css/main.css">

...

<div id="target" class="socialcounters">
  <span class="sc-total-count" rel="total-count">0</span>

  <a class="sc-service sc-facebook" rel="facebook-link">
    <span class="sc-logo">Facebook</span><span class="sc-count" rel="facebook-count">0</span>
  </a>
  <a class="sc-service sc-twitter" rel="twitter-link">
    <span class="sc-logo">Twitter</span><span class="sc-count" rel="twitter-count">0</span>
  </a>
  <a class="sc-service sc-google" rel="google-link">
    <span class="sc-logo">Google</span><span class="sc-count" rel="google-count">0</span>
  </a>
</div>

...

<script>
$('#target').socialcounters();
</script>
````

#### Options

 * `url`: specify an url to fetch data. By default, it will use the current url.
 * `callback`: specify a function to run after data comes back from app server.
 * `shorten`: determine whether the counters should be shortened to use "m" (million) or "k" (thousand) suffices.
 This feature is off by default.

#### Mappings

It is required to mark your elements with `rel="something"` for the plugin to fill data correctly. The default mappings include:

 * Links: `facebook-link`, `twitter-link`, `google-link`
 * Counts: `facebook-count`, `twitter-count`, `google-count`
 * Total (sum of all counts): `total-count`

## Deployment

The app has been pre-configured for simple pushing to Google App Engine and Heroku.

### Google App Engine

Create your account and everything (see [Google](https://cloud.google.com/appengine/docs/go/gettingstarted/uploading)'s
document). Setup the SDK on your computer (see [here](https://cloud.google.com/appengine/docs/go/gettingstarted/devenvironment))
then execute the deploy command:

````
goapp deploy
````

And you are done!

#### Further configuration

 * Depending on your need, you may want to enable dedicated memcache for improved performance.
 * Modify `app.yaml` and change the `WHITELIST` and/or `BLACKLIST` environment variables to secure
 and prevent other people from abusing your app.

### Heroku

Create your account, setup the [Heroku CLI](https://devcenter.heroku.com/articles/heroku-command) and do a push from your
forked repo:

````
git push heroku master
````

The app should be deployed without issue.

#### Further configuration

 * To make use of memory caching, enable the [Heroku MemCachier add-on](https://devcenter.heroku.com/articles/memcachier) for your application.
````
heroku addons:create memcachier
````
 * Add `WHITELIST` and/or `BLACKLIST` in Heroku Config Vars page to secure and prevent other people from abusing your app.
