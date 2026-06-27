/* csend site — motion.
   Deux effets, pas vingt : (1) le terminal du hero "imprime" une vraie session,
   (2) une séquence scroll signature fait descendre un paquet à travers les 8 couches.
   Tout dégrade proprement : sans GSAP ou en prefers-reduced-motion, rien ne casse. */
(function () {
  "use strict";
  var reduce = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  /* ---------- 1. Terminal du hero ---------- */
  var screen = document.getElementById("termScreen");
  var termLines = [
    '<span class="dim">$</span> <span class="p">csend</span> list',
    '<span class="ok">●</span> idle     SACEM       c9003b80   build prod',
    '<span class="p">◐</span> busy     ROYALEAI    ddb7d92f   backend',
    '<span class="dim">⚠ confirm JOLANANAS  784932b9   rm -rf ./dist (y/N)</span>',
    "",
    '<span class="dim">$</span> <span class="p">csend</span> inbox SACEM <span class="ok">"lance le build de prod"</span>',
    '<span class="ok">✓</span> déposé dans l’inbox de SACEM <span class="dim">[inbox]</span>',
    "",
    '<span class="dim">$</span> <span class="p">csend</span> recv',
    '1 message pour SACEM :',
    '  <span class="p">•</span> de ROYALEAI : <span class="ok">build vert, à toi ›</span>'
  ];

  function renderTerminal(animated) {
    if (!screen) return;
    screen.setAttribute("aria-hidden", "true");
    if (!animated) {
      screen.innerHTML = termLines.join("\n") + '<span class="cursor-block" style="height:1em"></span>';
      return;
    }
    var i = 0;
    screen.innerHTML = "";
    (function step() {
      if (i >= termLines.length) {
        screen.innerHTML = termLines.slice(0, i).join("\n") +
          '\n<span class="cursor-block" style="height:1em"></span>';
        return;
      }
      screen.innerHTML = termLines.slice(0, i + 1).join("\n") +
        '<span class="cursor-block" style="height:1em"></span>';
      i++;
      // empty lines flash by; real lines linger like a printing terminal
      var delay = termLines[i - 1] === "" ? 90 : 300;
      setTimeout(step, delay);
    })();
  }

  /* ---------- 2. Séquence scroll : le paquet traverse les couches ---------- */
  function initScroll() {
    var layers = Array.prototype.slice.call(document.querySelectorAll("#layers li"));
    var packet = document.getElementById("packet");
    var stack = document.querySelector(".stack");
    if (!layers.length) return;

    if (reduce || !window.gsap || !window.ScrollTrigger) {
      // fallback honnête : tout est lisible, mis en avant, sans paquet animé
      layers.forEach(function (l) { l.classList.add("is-active"); });
      return;
    }
    window.gsap.registerPlugin(window.ScrollTrigger);

    function place(idx) {
      var t = layers[idx];
      var y = t.offsetTop + t.offsetHeight / 2 - 13;
      window.gsap.to(packet, { y: y, duration: 0.25, overwrite: true, ease: "power2.out" });
      layers.forEach(function (l, i) { l.classList.toggle("is-active", i === idx); });
    }
    place(0);

    window.ScrollTrigger.create({
      trigger: stack,
      start: "top 62%",
      end: "bottom 72%",
      scrub: true,
      onUpdate: function (self) {
        var idx = Math.min(layers.length - 1, Math.floor(self.progress * layers.length));
        if (idx < 0) idx = 0;
        place(idx);
      }
    });
  }

  /* ---------- boot ---------- */
  renderTerminal(!reduce);
  if (document.readyState === "complete") {
    initScroll();
  } else {
    window.addEventListener("load", initScroll);
  }
})();
