// 迁移旧版本的 Service Worker。独立生图页必须每次经过服务端功能开关，
// 不再提供离线回退，避免已缓存页面绕过权限和功能门禁。
self.addEventListener('install', (event) => {
  event.waitUntil(self.skipWaiting())
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys()
      .then((keys) => Promise.all(keys.map((key) => caches.delete(key))))
      .then(() => self.registration.unregister()),
  )
})
