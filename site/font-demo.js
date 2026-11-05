const languageNames = { en: 'English', de: 'Deutsch', es: 'Español', fr: 'Français', ja: '日本語', zh: '简体中文' };
// A private CSS alias keeps the downloaded face out of the system-font comparison.
const bundledFamily = 'ntemoji Noto Color Emoji';

// A missing font-tech query is inconclusive: try the files in manifest order.
export function emojiFontCandidates(faces, css = globalThis.CSS) {
  const supports = css?.supports?.bind(css);
  const canQuery = supports?.('font-tech(variations)') === true;
  return faces.filter(face => !canQuery || supports(`font-tech(${face.tech})`));
}

export async function loadEmojiFace(faces, environment = globalThis) {
  const failures = [];
  for (const face of emojiFontCandidates(faces, environment.CSS)) {
    const url = new URL(`fonts/${face.file}`, import.meta.url);
    const font = new environment.FontFace(bundledFamily, `url("${url.href}")`, {
      weight: String(face.weight), style: face.style,
      // Leave Noto's oversized space to the text font.
      unicodeRange: 'U+0021-10FFFF',
    });
    try {
      await font.load();
      environment.document.fonts.add(font);
      return face;
    } catch (error) {
      failures.push(error); // A supported format may fail; try the next candidate.
    }
  }
  if (failures.length) throw new AggregateError(failures, 'Could not load any supported emoji font');
  return null;
}

export function setupFontDemo() {
  const select = document.getElementById('emoji-font');
  const language = document.getElementById('specimen-language');
  const status = document.getElementById('status');
  const samples = document.getElementById('emoji-specimens');
  const unsupportedMessage = 'Neither bundled emoji font format is supported here. Using system emoji.';
  let manifest, loadedFace, selected = 0;
  const buttons = [];

  function describe() {
    const locale = language.value;
    for (const [index, button] of buttons.entries()) {
      const specimen = manifest.specimens[index];
      const name = specimen.names[locale] || specimen.names.en;
      button.setAttribute('aria-label', name);
      button.setAttribute('aria-pressed', String(index === selected));
      button.lang = locale;
      button.querySelector('.specimen-caption').textContent = name;
    }
    const specimen = manifest.specimens[selected];
    document.getElementById('specimen-glyph').textContent = specimen.text;
    const name = document.getElementById('specimen-name');
    name.textContent = specimen.names[locale] || specimen.names.en;
    name.lang = locale;
    document.getElementById('specimen-codepoints').textContent = Array.from(specimen.text,
      character => `U+${character.codePointAt(0).toString(16).toUpperCase().padStart(4, '0')}`).join(' ');
  }

  async function changeFont() {
    select.disabled = true;
    status.dataset.state = 'loading';
    try {
      if (select.value === 'noto') {
        status.textContent = 'Loading Noto Color Emoji…';
        if (!loadedFace) loadedFace = await loadEmojiFace(manifest.faces);
        if (!loadedFace) {
          select.value = 'system';
          status.textContent = unsupportedMessage;
          status.dataset.state = 'ready';
          return;
        }
        const family = JSON.stringify(bundledFamily);
        // An explicit platform emoji fallback makes Safari prefer it for VS16
        // and keycap sequences. Request this face alone; missing glyphs can
        // still use the browser's implicit fallback.
        document.documentElement.style.setProperty('--demo-emoji-font', family);
        status.textContent = `Noto Color Emoji requested · ${loadedFace.tech.replace('color-', '')} · ${(loadedFace.bytes / 1000000).toFixed(1)} MB loaded. Browser fallback may apply.`;
      } else {
        document.documentElement.style.removeProperty('--demo-emoji-font');
        status.textContent = loadedFace ? 'System emoji · Using your platform’s default appearance.' : 'System emoji · No font download.';
      }
      status.dataset.state = 'ready';
    } catch (error) {
      select.value = 'system';
      document.documentElement.style.removeProperty('--demo-emoji-font');
      status.textContent = 'Could not load the available Noto fonts. Using system emoji; check your connection and select Noto to retry.';
      status.dataset.state = 'error';
      console.error(error);
    } finally {
      select.disabled = false;
    }
  }

  language.addEventListener('change', describe);
  select.addEventListener('change', changeFont);
  (async () => {
    const response = await fetch(new URL('fonts/specimens.json', import.meta.url));
    if (!response.ok) throw new Error(`Sequence samples failed to load: ${response.status}`);
    manifest = await response.json();
    for (const locale of manifest.locales) language.add(new Option(languageNames[locale] || locale, locale));
    const preferred = new URLSearchParams(location.search).get('lang');
    language.value = manifest.locales.includes(preferred) ? preferred : 'en';
    for (const [index, specimen] of manifest.specimens.entries()) {
      const button = document.createElement('button');
      button.type = 'button';
      const glyph = document.createElement('span');
      glyph.className = 'specimen-emoji';
      glyph.setAttribute('aria-hidden', 'true');
      glyph.textContent = specimen.text;
      const caption = document.createElement('span');
      caption.className = 'specimen-caption';
      button.append(glyph, caption);
      button.addEventListener('click', () => { selected = index; describe(); });
      buttons.push(button);
      samples.append(button);
    }
    describe();
    language.disabled = false;
    select.disabled = false;
    const unsupported = emojiFontCandidates(manifest.faces).length === 0;
    const option = select.querySelector('[value="noto"]');
    option.disabled = unsupported;
    if (unsupported) option.textContent = 'Noto Color Emoji · unsupported here';
    status.textContent = unsupported ? unsupportedMessage : 'System emoji · No font download.';
    status.dataset.state = 'ready';
  })().catch(error => {
    status.textContent = 'Could not load the font comparison. Reload to try again.';
    status.dataset.state = 'error';
    console.error(error);
  });
}
