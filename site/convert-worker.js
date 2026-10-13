// The only conversion engine is the Go library. Keep it off the UI thread.
importScripts('./wasm_exec.js');
const ready = new Promise((resolve, reject) => {
  self.ntemojiReady = () => resolve(self.ntemoji);
  (async () => {
    const go = new Go();
    const response = await fetch('./demos/convert.wasm');
    if (!response.ok) throw new Error(`WASM download failed: ${response.status}`);
    const { instance } = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject);
    go.run(instance).then(() => reject(new Error('Conversion engine exited')), reject);
  })().catch(reject);
});
ready.then(api => postMessage({ type: 'ready', locales: api.locales }))
  .catch(error => postMessage({ type: 'fatal', error: error.message }));
self.onmessage = async ({ data }) => {
  try {
    const api = await ready;
    if (typeof data.text !== 'string' || data.text.length > 100000) throw new Error('Input is limited to 100,000 characters.');
    postMessage({ id: data.id, ...api.convert(data) });
  } catch (error) {
    postMessage({ id: data.id, error: error.message });
  }
};
