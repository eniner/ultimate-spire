const IDB_NAME = "keyval-store"
const IDB_STORE = "keyval"
const IDB_KEY = "eqdir"

function openKeyval(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(IDB_NAME, 1)
    req.onupgradeneeded = () => {
      const db = req.result
      if (!db.objectStoreNames.contains(IDB_STORE)) {
        db.createObjectStore(IDB_STORE)
      }
    }
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error)
  })
}

export async function readEqDirHandle(): Promise<any> {
  const db = await openKeyval()
  try {
    return await new Promise((resolve, reject) => {
      const tx = db.transaction(IDB_STORE, "readonly")
      const req = tx.objectStore(IDB_STORE).get(IDB_KEY)
      req.onsuccess = () => resolve(req.result || null)
      req.onerror = () => reject(req.error)
    })
  } finally {
    db.close()
  }
}

export async function writeEqDirHandle(handle: any): Promise<void> {
  const db = await openKeyval()
  try {
    await new Promise<void>((resolve, reject) => {
      const tx = db.transaction(IDB_STORE, "readwrite")
      tx.objectStore(IDB_STORE).put(handle, IDB_KEY)
      tx.oncomplete = () => resolve()
      tx.onerror = () => reject(tx.error)
    })
  } finally {
    db.close()
  }
}

export async function eqDirGranted(handle: any): Promise<boolean> {
  if (!handle || typeof handle.queryPermission !== "function") {
    return false
  }
  try {
    return (await handle.queryPermission({mode: "readwrite"})) === "granted"
  } catch (_err) {
    return false
  }
}

export async function grantEqDir(handle: any): Promise<boolean> {
  if (!handle || typeof handle.requestPermission !== "function") {
    return false
  }
  try {
    return (await handle.requestPermission({mode: "readwrite"})) === "granted"
  } catch (_err) {
    return false
  }
}

export async function zoneLooksConverted(handle: any, zone: string): Promise<boolean> {
  if (!handle || !zone) {
    return false
  }
  const name = String(zone).toLowerCase()
  try {
    const sage = await handle.getDirectoryHandle("eqsage")
    try {
      await sage.getDirectoryHandle(name)
      return true
    } catch (_err) {
      // keep looking
    }
    for await (const [entryName] of sage.entries()) {
      if (String(entryName).toLowerCase().indexOf(name) !== -1) {
        return true
      }
    }
  } catch (_err) {
    return false
  }
  return false
}

export async function connectEqDirectory(forcePick = false): Promise<{handle: any; granted: boolean; error: string}> {
  if (typeof (window as any).showDirectoryPicker !== "function") {
    return {handle: null, granted: false, error: "This browser cannot pick a folder. Open Sage in Chrome, then use Connect EQ folder."}
  }
  let handle = await readEqDirHandle()
  if (handle && !forcePick && !(await eqDirGranted(handle))) {
    const granted = await grantEqDir(handle)
    if (granted) {
      await writeEqDirHandle(handle)
      return {handle, granted: true, error: ""}
    }
  }
  if (!handle || forcePick || !(await eqDirGranted(handle))) {
    try {
      handle = await (window as any).showDirectoryPicker({mode: "readwrite"})
    } catch (err: any) {
      if (err && err.name === "AbortError") {
        return {handle: null, granted: false, error: ""}
      }
      return {handle: null, granted: false, error: (err && err.message) || "Folder picker failed."}
    }
  }
  if (!handle || handle.kind !== "directory") {
    return {handle: null, granted: false, error: "That was not a folder."}
  }
  const granted = await grantEqDir(handle)
  if (!granted) {
    return {handle, granted: false, error: "Chrome blocked write access. Click Allow so Sage can write the eqsage cache next to your EQ files."}
  }
  await writeEqDirHandle(handle)
  return {handle, granted: true, error: ""}
}
