No Noise Please
===============

What
----

[Demo](https://nonoiseplease.com)

nonoiseplease is where i would like to save my bookmarks  
It lets you "index" pages from the web and makes them searchable through a full-text search engine.  

Why
---

I have always thought that every day I come across sections of the web that are particularly interesting, and I've always tried to save them for later use... in my bookmarks. However, the "later use" never actually came, and I kept thinking, "Damn, there must be some hidden gems buried deep in my browser bookmarks."  
So, I decided to create a tool that could scrape and index my own web findings. Then, I realized that this tool could be useful for others as well. After all, I can't be the only one who still uses bookmarks to save everything in an unsorted manner.  
Also, Google sometimes is shit and SEO is a bit annoying, that makes retrieving some niche page I find while surfing very difficult.  
Also, AI will generate a ton of noise on the internet and i want to try to increase "signal" / "shit" ratio on my searches.  

How
---

#### To use

*   Create an account: You can use a **temporary email service** like Temp-Mail or a tool like Firefox Relay. It's fine if you choose either of those options; I just needed a way to differentiate users and their indexes.
*   Log in: Enter your credentials to access your account.
*   Add URLs to scrape: Provide the software with a URL to scrape. I understand that it might be considered unethical to take content from others, but I couldn't think of a better approach. The scraped content will be added to your index.
*   Repeat the process: You can repeat steps 3 multiple times to add more content to your index. However, please note that this functionality is currently limited due to the risk of IP banning or blacklisting.
*   Alternatively, use the Firefox extension: I have created a Firefox extension that allows you to upload your current page directly to your index. This extension is useful because it has no limits, and it can upload personal data displayed on dynamic pages. However, please be cautious as I am still a rando person on the internet.
*   Manage your pages: You can add categories to your pages in the "/me/pages" section
*   Search for pages: Visit the provided website and use the search function with "keywords" (I mean... the query is sent directly to MeiliSearch). You will receive a list of pages from your index that match the query.

#### It's done

On the backend: Golang (PocketBase), along with MeiliSearch for the index (every user has his own personal index)  
On the frontend: MVP.css (a little adapted), Alpine.js, js-cookie and some Golang templating.  
Tests are done with Python  

Limitations
-----------

Right now, the main limitation is on the number of pages that can be scraped. I have set a limit of 5 pages per month.  
This is because I don't want to get my ip blacklisted. As a makeshift, I am currently routing the requests through a proxy, but I am still working on a solution to this problem. Suggestion are welcome.  
Also, I call "scaping" a simple GET request to the page (and so the text), so if you need to login to see the content, you can't scrape it (but you can use the extension).  

Privacy
-------

Well, of course, if I want to index every web page uploaded, I think there is no alternative but to obtain the clear/plain text (note that this is also valid if you want to index personal data through the extension). In case there is a way to protect the indexed text as well, please let me know.  
I ask only for an email and save only 2 cookies (that a know of): a JWT and an id.  
There are no analitycs.  
The extention is 1.6kB of js, you can look it up. (just rename the file from .xpi to .zip)  

Alternatives
---
With what i see as pros or cons.

### [Grimoire](https://github.com/goniszewski/grimoire)
##### Pros
- add your personal notes to bookmarks
##### Cons
- has to be self-hosted

### [Linkwarden](linkwarden.app)
##### Pros
- stores the page
##### Cons
- free only if self-hosted
- a little "cumbersome" when adding new links

### [Readclip](https://readclip.site/)
##### Pros
- stores the page
- easy bookmark add
##### Cons
- no source

### [Raindrop](https://raindrop.io)
##### Pros
- full optional
##### Cons
- requires apps

### [xbrowsersync](https://github.com/xbrowsersync)
##### Pros
- free, open source, anonymous
#### Cons
- requires installation

### [Zotero](https://www.zotero.org/)
##### Pros
- open source
##### Cons
- too much focused on research (has features i m not interested in)

### [Pinboard](pinboard.in)
##### Pros
- simple, fast, whit API (perferct)
##### Cons
- only paid version

### [del.icio.us](https://del.icio.us)
- maybe somethong to aspire to

Future
------

nonoiseplease is still missing some features, I know, but I'm working on it.  

*   browser extention
    *   fix the terrible UX and auth
    *   add in batch all of your bookmark (i really want it cause i m lazy but i dont know how to do without massive scraping)
    *   show search result (maybe along side google)
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