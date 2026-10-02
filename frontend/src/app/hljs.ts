import hljs from "highlight.js/lib/core"
import json from "highlight.js/lib/languages/json"

// Register only the grammar this app actually highlights. The 9.x ReDoS
// advisory (GHSA-7wwv-vh3v-89cq) lives in other bundled grammars; keeping
// the core import means those grammars are never loaded.
hljs.registerLanguage("json", json)

export default hljs
