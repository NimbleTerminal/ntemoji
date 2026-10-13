const demos = {
  demoji: {
    title: 'Demoji · emoji to words',
    description: 'Type or paste emoji text and see descriptive names appear live. Choose a language and plain names, colons, brackets, or your own delimiters.',
  },
  remoji: {
    title: 'Remoji · names to emoji',
    description: 'Convert names such as :rocket: to emoji, or choose Plain names for greedy matching. Choose a language, then load an example to try localized names. Unknown names stay unchanged.',
  },
  simple: {
    title: 'Emoji picker',
    description: 'Browse categories, try the presets, or Tab to search. This picker uses a custom forest theme. Selecting an emoji ends the demo; restart to choose another.',
  },
  swatch: {
    title: 'Swatch grid',
    description: 'Click a swatch, or press 1–4, to open its picker. Choose an emoji and see the button update. Press q outside a picker to finish.',
  },
};
const requested = new URLSearchParams(location.search).get('demo');
const selected = Object.hasOwn(demos, requested) ? requested : 'swatch';
const demo = demos[selected];
const status = document.querySelector('#status');
let terminal;
let failed = false;
const conversion = selected === 'demoji' || selected === 'remoji';

function setStatus(message, state = 'loading') {
  status.textContent = message;
  status.dataset.state = state;
}
function fail(error) {
  failed = true;
  setStatus('Could not run the demo. Check your connection and reload the page.', 'error');
  console.error(error);
}

document.title = `${demo.title} — ntemoji`;
document.querySelector('#demo-title').textContent = demo.title;
document.querySelector('#description').textContent = demo.description;
document.querySelector('#source').href = `https://github.com/NimbleTerminal/ntemoji/tree/main/${conversion ? "cmd/emojiwasm" : `examples/${selected}`}`;
document.querySelectorAll('[data-demo]').forEach(link => {
  if (link.dataset.demo === selected) link.setAttribute('aria-current', 'page');
  else link.removeAttribute('aria-current');
});
// Full navigation gives each demo a fresh Go runtime and releases the old one.
document.querySelector('#restart').addEventListener('click', () => location.reload());
window.addEventListener('pagehide', () => terminal?.dispose());

async function main() {
  if (conversion) {
    document.querySelector('.intro h1').innerHTML = selected === 'demoji'
      ? 'Emoji to words. In your language<span>.</span>'
      : 'Words back to emoji<span>.</span>';
    document.querySelector('.intro > p:last-child').textContent = 'A text converter in six languages. Edit, compare, and copy — all on your device.';
    document.querySelector('#converter').hidden = false;
    document.querySelector('#controls').hidden = true;
    document.querySelector('.terminal-scroll').hidden = true;
    document.querySelector('#keyboard-help').hidden = true;
    document.querySelector('.instructions .hint').textContent = 'Choose settings above, then type or paste text. Conversion runs locally; your text is never uploaded.';
    document.querySelector('#runtime-label').textContent = 'Go library · WebAssembly';
    document.querySelector('#restart').textContent = 'Reload demo';
    const { startConverter } = await import('./converter.js');
    startConverter(selected, setStatus, fail);
    return;
  }
  const { BoobaTerminal, installActionBar } = await import('./booba/booba.js');
  // Conversion pages load this runtime only inside their worker.
  await import('./wasm_exec.js');
  const go = new Go();
  // Buffering also works on static servers without an application/wasm MIME type.
  const response = await fetch(`demos/${selected}.wasm`);
  if (!response.ok) throw new Error(`Demo download failed: ${response.status}`);
  const { instance } = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject);
  terminal = new BoobaTerminal('terminal', { fontSize: 15, allowOSC7501Actions: true, renderer: 'canvas2d' });
  await terminal.init();
  terminal.setActionKeyEncoder({
    canEncode: gesture => terminal.encodeKey(gesture) !== null,
    encode: gesture => terminal.encodeKey(gesture),
  });
  installActionBar(terminal.actions, {
    parent: document.querySelector('#controls'),
    openKeyboard: () => terminal.focus(),
  });
  terminal.onActionError = error => {
    setStatus('Could not send that action · Try again or use the keyboard', 'error');
    console.error(error);
  };
  terminal.onStatusChange = (state, message) => {
    if (failed) return;
    if (state === 'connected') setStatus('Ready · Click the terminal to start', 'ready');
    else if (state === 'disconnected') {
      if (message === 'Program exited') setStatus('Demo finished · Restart to try again', 'finished');
      else setStatus('Demo disconnected · Restart to try again', 'error');
    }
  };
  go.run(instance).catch(fail);
  terminal.connectWasm();
}
main().catch(fail);
