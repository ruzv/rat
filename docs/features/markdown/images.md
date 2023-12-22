---
id: 70f28579-9bc2-4134-b27c-8a5987831f55
---

# images

<rat graph depth=1 />

---

Rat fully supports markdown image syntax - `![alt text](url/to/image)`.

Additionally, if file servers are configured a relative path can be provided,
pointing to an image served by one of the file servers. And Rat will
automatically handle auth to the server and serve the image.

Example:

Let's say we have a file server providing the following image
`https://fileserver/img/dodle.png`, then the Markdown syntax to render the image
would be `![alt text](/img/dodle.png)`. Rat will resolve the URL to
`https://rat/graph/file/img/dodle.png` and proxy the contents.
