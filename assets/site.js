// Copy-to-clipboard for any [data-copy] element. Falls back to text selection
// if the Clipboard API is unavailable (older Safari, http contexts).
(function () {
    function flash(el) {
        el.classList.add("copied");
        var prev = el.querySelector(".cmd")?.textContent;
        var label = el.querySelector(".cmd");
        if (label) label.textContent = "copied";
        setTimeout(function () {
            el.classList.remove("copied");
            if (label && prev) label.textContent = prev;
        }, 1200);
    }

    function copy(text, el) {
        if (navigator.clipboard && window.isSecureContext) {
            navigator.clipboard.writeText(text).then(
                function () {
                    flash(el);
                },
                function () {
                    legacyCopy(text, el);
                },
            );
        } else {
            legacyCopy(text, el);
        }
    }

    function legacyCopy(text, el) {
        var ta = document.createElement("textarea");
        ta.value = text;
        ta.style.position = "fixed";
        ta.style.opacity = "0";
        document.body.appendChild(ta);
        ta.select();
        try {
            document.execCommand("copy");
            flash(el);
        } catch (_) {}
        document.body.removeChild(ta);
    }

    document.addEventListener("click", function (e) {
        var el = e.target.closest("[data-copy]");
        if (!el) return;
        e.preventDefault();
        copy(el.getAttribute("data-copy"), el);
    });
})();
