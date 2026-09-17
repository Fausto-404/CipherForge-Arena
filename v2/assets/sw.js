self.addEventListener('install', (event) => event.waitUntil(self.skipWaiting()));
self.addEventListener('activate', (event) => event.waitUntil(self.clients.claim()));
self.addEventListener('fetch', (event) => {
  if (new URL(event.request.url).pathname.endsWith('/v2/api.php')) {
    event.respondWith(fetch(event.request, {headers: new Headers(event.request.headers)}));
  }
});
