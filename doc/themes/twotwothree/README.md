onetwothree
===========

This is a  minimalistic, responsive theme that is simple as *one, two, three.* 
It has some neat features like captioning images, but its best quality is its lack of features (no jQuery, no huge image banners, no post previews, no word counts, etc.) which I've always found unnecessary and thus cluttering.

Check out the demo at https://schollz.github.io/onetwothree/.

![Screenshot of theme](https://raw.github.com/schollz/onetwothree/master/images/screenshot.png)


# Using

```
git clone https://github.com/schollz/onetwothree.git themes/onetwothree
hugo server -t onetwothree
```

# Configuration

Add these parameters to your configuration file:

```
[params]
    twitter = "yakczar"
    navigation = ["about.md"]
```

The `twitter` parameter will put a footer on your pages on how to contact you (with the subject being the page URL) to comment on the page.

The `navigation` parameter will pick which pages to list at the top bar. I'm aware that the template HTML could just loop through the pages, but I've found that this method is 120x faster, which is useful when you have thousands of posts.

# License

MIT
