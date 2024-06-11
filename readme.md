No Noise Please
===============

What
----

[Demo](https://nonoiseplease.com)

nonoiseplease is an experiment to try an "enahnce" my web search results.
It lets you "index" pages from the web and makes them searchable through a full-text search engine.  
Eventually, with an extension, it can show indexed pages onto google results.

Why
---

I have always thought that every day I come across sections of the web that are particularly interesting, and I've always tried to save them for later use... in my bookmarks. However, the "later use" never actually came, and I kept thinking, "Damn, there must be some hidden gems buried deep in my browser bookmarks."  
So, I decided to create a tool that could scrape and index my own web findings. Then, I realized that this tool could be useful for others as well. After all, I can't be the only one who still uses bookmarks to save everything in an unsorted manner.  
Also, Google sometimes is shit and SEO is a bit annoying, that makes retrieving some niche page I find while surfing very difficult.  
Also, AI will generate a ton of noise on the internet and i want to try to increase "signal" / "shit" ratio on my searches.  

## How To use

(please can you read on the website? i was repeating/fragmenting to much information, sorry)

## How it works

3 main components
- backend
- meilisearc
- extension

## backend
let s split in actual backend (pocketbase) and frontend
#### PocketBase
god bless pocketbase, the right spot between simple, complex, extendible... also, coming from i bit of python, i really needed types
- it is still a giant mess (i cant write idiomatic go)
	- i have a set of controller with their interfaces
	- every api call is handled by WebController (NoNoiseInterface) and i listed in be.go
		-  WebController handle everything that is not handled by PocketBase exposed api
			- mainly scrape, search, categorize pages, communication with extension
	- pb_public and extension_template has ui
	- serve_public handles templating
- luckly pocketbase has a giant set of hooks that i use, for example, to delete every page a user scraped when his record is deleted
#### Frontend
i have no idea on who to do frontend, it s terrible
- there are a bunch of pages just for "info" (/why, /index, /alternatives)
- pages with more user interaction are a mix of go-templating and alpine-framework
	- templated pages are me/account.html and register.html, are listed on backend/serve_public/template_renderer.go
		- i like templating cause it can always be usefull, also the HTML sent to extension is templated (backend/extention_template/search.html)
	- most is done with alpine cause the backend expose some api
		- authenticated with jwt  
- pages enables all function all the backend
	- CRUD user
	- scrape a webpage
	- search/catogory scraped web pages

## meilisearch
is the engine that provide full text search
- it comunicates only with the backend
	- when a user registes, it gets its own index
	- the index is kept synched at every operation
- every index stores the docs as full text, list of categories and id (to have a link ok the backend)
	- it s the only element that stores the full text scraped

## extension
actually it s optional, it s just to avoid scraping from backend ip and serving results inside google, not much to explain
- inject HTML (dangerous) on google search, with results from a search on backend
	- when you search on google, the extensione recognize that and does the same search on backend, which respondes with html
- allows you to upload you current page to backend
	- simply by reading the source
- eventually:
	- records the pages you visit locally (in background script) and send them to backend if you notice you found something interesting
 	- upload your bookmarks. they get buffered and patiently (slowly) scraped 
- saves cookies to remember your options

Not Released
------------

I have 
- some playwright (and other python) to test the website. Currently i dont test the extension.
- some data i test and develop on
- some shell to build, deploy, rollback almost autocatically
But these are *definetly* not ready to be shared publicly


Limitations
-----------

Right now, the main limitation is on the number of pages that can be scraped. I have set a limit of 5 pages per month.  
This is because I don't want to get my ip blacklisted. As a makeshift, I am currently routing the requests through a proxy, but I am still working on a solution to this problem. Suggestion are welcome.  
Also, I call "scaping" a simple GET request to the page (and so the text), so if you need to login to see the content, you can't scrape it (but you can use the extension).  

Privacy
-------

Well, of course, if I want to index every web page uploaded, I think there is no alternative but to obtain the clear/plain text (note that this is also valid if you want to index personal data through the extension). In case there is a way to protect the indexed text as well, please let me know. Maybe, indexing can be optional.. dunno.
I ask only for an email and save only 2 cookies (that a know of): a JWT and an id.  
There are no analitycs.  
If you delete your account, it is actually deleted.  
The extention is 1.6kB of js, you can look it up. (just rename the file from .xpi to .zip)  




Future
------

nonoiseplease is still missing some features, I know, but I'm working on it.  

*   browser extention
    *   fix the terrible UX and auth
*   improve the "scraper"
    *   to get the transcript of a Youtube video
    *   find a way to reliably read PDF
    *   consider to add something like puppeteer/playwright
*   make a clear rest API
    *   you can alerady try to reverse engeer the js (should not be too difficult)
*   (in a far future) give you the option to make your personal index searchable by others
    *   cause maybe you are the top expert on a very little niche and you found interesting resources (that other may find interesting)
    *   with a clear distinction of what you are sharing / where are your searches going
    *   maybe also add a way to "group" pages (like a tag, but with a description)
*   got to add a local client
    *   maybe gui (https://mattn.github.io/go-gtk/), maybe not


---

<h2>
            Alternatives
        </h2>
        <div>
            <h3><a href="https://github.com/goniszewski/grimoire">Grimoire</a></h3>
        <h4>Pros</h4>
        - add your personal notes to bookmarks
        <h4>Cons</h4>
        - has to be self-hosted
        </div>
        <div>
            <h3><a href="linkwarden.app">Linkwarden</a></h3>
        <h4>Pros</h4>
        - stores the page
        <h4>Cons</h4>
        - free only if self-hosted
        - a little "cumbersome" when adding new links
        </div>
        <div>
            <h3><a href="https://readclip.site">Readclip</a></h3>
        <h4>Pros</h4>
        - stores the page
        - easy bookmark add
        <h4>Cons</h4>
        - no source
        </div>
        <div>
            <h3><a href="https://raindrop.io">Raindrop</a></h3>
        <h4>Pros</h4>
        - full optional
        <h4>Cons</h4>
        - requires apps
        </div>
        <div>
            <h3><a href="https://github.com/xbrowsersync">xbrowsersync</a></h3>
        <h4>Pros</h4>
        - free, open source, anonymous
        <h4>Cons</h4>
        - requires installation
        </div>
        <div>
            <h3><a href="https://www.zotero.org">Zotero</a></h3>
        <h4>Pros</h4>
        - open source
        <h4>Cons</h4>
        - too much focused on research (has features i m not interested in)
        </div>
        <div>
            <h3><a href="pinboard.in">Pinboard</a></h3>
        <h4>Pros</h4>
        - simple, fast, whit API (perferct)
        <h4>Cons</h4>
        - only paid version
        </div>
        <div>
            <h3><a href="https://del.icio.us">del.icio.us</a></h3>
            <span>✨maybe somethong to aspire to✨</span>
        </div>

Screenshot Dump
---------------

Home
![home](https://github.com/byschii/nonoiseplease/blob/master/screenshots/1main.png)

Browser Extension
<br />
![Extension](https://github.com/byschii/nonoiseplease/blob/master/screenshots/2ext.png)

In-google search (for extension)
![googlesearch](https://github.com/byschii/nonoiseplease/blob/master/screenshots/7gs.png)

Account
![Account](https://github.com/byschii/nonoiseplease/blob/master/screenshots/3acc.png)

Scraped Pages
![Scraped Pages](https://github.com/byschii/nonoiseplease/blob/master/screenshots/4pages.png)

Page Details
![details](https://github.com/byschii/nonoiseplease/blob/master/screenshots/5detail.png)

Search
![search](https://github.com/byschii/nonoiseplease/blob/master/screenshots/6search.png)



