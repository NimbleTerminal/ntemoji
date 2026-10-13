const names = { en: 'English', de: 'Deutsch', es: 'Español', fr: 'Français', ja: '日本語', zh: '简体中文' };
const example = 'Ready to launch 🚀\nGreat work 👍 🎉\nExplorer: 🧑🏽‍🚀';

export function startConverter(mode, setStatus, fail) {
  const $ = id => document.getElementById(id);
  const source = $('source-text'), result = $('result-text');
  const language = $('language'), format = $('format'), copy = $('copy-result');
  const worker = new Worker('./convert-worker.js');
  let version = 0, timer, sampleID = null, ready = false;
  format.value = mode === 'demoji' ? 'plain' : 'colons';
  const delimiters = () => {
    if (format.value === 'plain') return ['', ''];
    if (format.value === 'colons') return [':', ':'];
    if (format.value === 'brackets') return ['[', ']'];
    return [$('opening').value, $('closing').value];
  };
  function request(text, requestMode = mode, id = version) {
    const [open, close] = delimiters();
    worker.postMessage({ id, mode: requestMode, text, locale: language.value, open, close });
  }
  function update() {
    version++;
    sampleID = null;
    clearTimeout(timer);
    copy.disabled = true;
    $('custom-delimiters').hidden = format.value !== 'custom';
    const [open, close] = delimiters();
    $('converter-hint').textContent = mode === 'remoji' && (!open || !close)
      ? 'Plain names use greedy whole-name matching and can replace ordinary words. Input stays on this device.'
      : 'Text stays on this device. Language changes preserve your input.';
    const url = new URL(location.href);
    url.searchParams.set('lang', language.value);
    history.replaceState(null, '', url);
    result.lang = mode === 'demoji' ? language.value : '';
    setStatus('Converting…');
    timer = setTimeout(() => request(source.value), 120);
  }
  function loadExample() {
    clearTimeout(timer);
    version++;
    sampleID = null;
    if (mode === 'demoji') { source.value = example; update(); return; }
    copy.disabled = true;
    sampleID = version;
    setStatus('Preparing localized example…');
    request(example, 'demoji');
  }
  function fatal(error) {
    ready = false;
    clearTimeout(timer);
    $('converter-fields').disabled = true;
    worker.terminate();
    fail(error);
  }
  worker.onmessage = ({ data }) => {
    if (data.type === 'fatal') { fatal(new Error(data.error)); return; }
    if (data.type === 'ready') {
      for (const locale of data.locales) language.add(new Option(names[locale] || locale, locale));
      const preferred = new URLSearchParams(location.search).get('lang');
      language.value = data.locales.includes(preferred) ? preferred : 'en';
      ready = true;
      $('converter-fields').disabled = false;
      loadExample();
      return;
    }
    if (data.id !== version) return; // Never display stale input/settings results.
    if (data.error) { result.value = ''; copy.disabled = true; setStatus(data.error, 'error'); return; }
    if (sampleID === data.id) {
      sampleID = null;
      source.value = data.text;
      update();
      return;
    }
    result.value = data.text;
    copy.disabled = result.value.length === 0;
    setStatus(source.value ? 'Converted locally · Ready to copy' : 'Ready · Type or paste text to begin', 'ready');
  };
  worker.onerror = event => fatal(new Error(event.message || 'Conversion worker failed'));
  for (const element of [source, language, format, $('opening'), $('closing')]) {
    element.addEventListener(element.tagName === 'SELECT' ? 'change' : 'input', () => { if (ready) update(); });
  }
  $('sample').addEventListener('click', loadExample);
  $('clear-text').addEventListener('click', () => { source.value = ''; update(); source.focus(); });
  copy.addEventListener('click', async () => {
    try { await navigator.clipboard.writeText(result.value); setStatus('Output copied', 'ready'); }
    catch { result.focus(); result.select(); setStatus('Select Copy from your browser, or press Ctrl+C / ⌘C.', 'ready'); }
  });
  window.addEventListener('pagehide', () => { clearTimeout(timer); worker.terminate(); }, { once: true });
}
