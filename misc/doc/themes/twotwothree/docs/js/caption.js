function ready(fn) {
    if (document.attachEvent ? document.readyState === "complete" : document.readyState !== "loading") {
        var elements = document.querySelectorAll("img");
        Array.prototype.forEach.call(elements, function(el, i) {
            if (el.getAttribute("alt")) {
                const caption = document.createElement('figcaption');
                var node = document.createTextNode(el.getAttribute("alt"));
                caption.appendChild(node);
                const wrapper = document.createElement('figure');
                wrapper.className = 'image';
                el.parentNode.insertBefore(wrapper, el);
                el.parentNode.removeChild(el);
                wrapper.appendChild(el);
                wrapper.appendChild(caption);
                console.log(el);
                console.log(el.outerHTML)
            }
        });
    } else {
        document.addEventListener('DOMContentLoaded', fn);
    }
}
window.onload = ready;
