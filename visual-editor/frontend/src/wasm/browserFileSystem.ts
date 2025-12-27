/**
 * Browser File System Abstraction
 *
 * Provides dual-mode file storage for Layer WASM:
 * - Virtual Mode: Uses localStorage (default, works everywhere)
 * - Real Mode: Uses File System Access API (Chrome/Edge only)
 */

export type StorageMode = 'virtual' | 'real';

export interface OptionResult {
  type: 'Some' | 'None';
  value?: any;
}

class BrowserFileSystem {
  private mode: StorageMode = 'virtual';
  private directoryHandle: FileSystemDirectoryHandle | null = null;
  private virtualPrefix = 'layerfs:';

  /**
   * Check if File System Access API is supported
   */
  isFileSystemAccessSupported(): boolean {
    return 'showDirectoryPicker' in window;
  }

  /**
   * Get current storage mode
   */
  getMode(): StorageMode {
    return this.mode;
  }

  /**
   * Set storage mode
   */
  setMode(mode: StorageMode): void {
    this.mode = mode;
    console.log(`[BrowserFS] Storage mode set to: ${mode}`);
  }

  /**
   * Request directory access from user
   * Shows native directory picker dialog
   */
  async requestDirectoryAccess(): Promise<boolean> {
    if (!this.isFileSystemAccessSupported()) {
      console.warn('[BrowserFS] File System Access API not supported in this browser');
      return false;
    }

    try {
      this.directoryHandle = await (window as any).showDirectoryPicker({
        mode: 'readwrite',
      });
      this.mode = 'real';
      console.log('[BrowserFS] Directory access granted');
      return true;
    } catch (err) {
      if ((err as Error).name !== 'AbortError') {
        console.error('[BrowserFS] Error requesting directory access:', err);
      } else {
        console.log('[BrowserFS] User cancelled directory picker');
      }
      return false;
    }
  }

  /**
   * Read file contents
   */
  async readFile(path: string): Promise<OptionResult> {
    // Try real file system first if in real mode
    if (this.mode === 'real' && this.directoryHandle) {
      try {
        const fileHandle = await this.getFileHandle(path);
        const file = await fileHandle.getFile();
        const content = await file.text();
        console.log(`[BrowserFS] Read file (real): ${path}`);
        return { type: 'Some', value: content };
      } catch (err) {
        console.warn(`[BrowserFS] Real FS read failed for ${path}:`, err);
        // Fall through to virtual mode
      }
    }

    // Virtual mode (localStorage)
    const key = this.virtualPrefix + path;
    const content = localStorage.getItem(key);

    if (content !== null) {
      console.log(`[BrowserFS] Read file (virtual): ${path}`);
      return { type: 'Some', value: content };
    }

    return { type: 'None' };
  }

  /**
   * Write file contents
   */
  async writeFile(path: string, content: string): Promise<OptionResult> {
    // Try real file system if in real mode
    if (this.mode === 'real' && this.directoryHandle) {
      try {
        const fileHandle = await this.getFileHandle(path, { create: true });
        const writable = await fileHandle.createWritable();
        await writable.write(content);
        await writable.close();
        console.log(`[BrowserFS] Wrote file (real): ${path}`);
        return { type: 'Some', value: content };
      } catch (err) {
        console.error(`[BrowserFS] Real FS write failed for ${path}:`, err);
        return { type: 'None' };
      }
    }

    // Virtual mode (localStorage)
    try {
      const key = this.virtualPrefix + path;
      localStorage.setItem(key, content);
      console.log(`[BrowserFS] Wrote file (virtual): ${path}`);
      return { type: 'Some', value: content };
    } catch (err) {
      console.error(`[BrowserFS] localStorage write failed for ${path}:`, err);
      return { type: 'None' };
    }
  }

  /**
   * Check if file exists
   */
  async exists(path: string): Promise<boolean> {
    // Try real file system if in real mode
    if (this.mode === 'real' && this.directoryHandle) {
      try {
        await this.getFileHandle(path);
        return true;
      } catch {
        return false;
      }
    }

    // Virtual mode
    const key = this.virtualPrefix + path;
    return localStorage.getItem(key) !== null;
  }

  /**
   * Delete file
   */
  async deleteFile(path: string): Promise<OptionResult> {
    // Try real file system if in real mode
    if (this.mode === 'real' && this.directoryHandle) {
      try {
        const parts = path.split('/').filter(p => p);
        const fileName = parts.pop()!;

        let currentDir = this.directoryHandle;
        for (const part of parts) {
          currentDir = await currentDir.getDirectoryHandle(part);
        }

        await currentDir.removeEntry(fileName);
        console.log(`[BrowserFS] Deleted file (real): ${path}`);
        return { type: 'Some', value: path };
      } catch (err) {
        console.warn(`[BrowserFS] Real FS delete failed for ${path}:`, err);
        return { type: 'None' };
      }
    }

    // Virtual mode
    const key = this.virtualPrefix + path;
    if (localStorage.getItem(key) !== null) {
      localStorage.removeItem(key);
      console.log(`[BrowserFS] Deleted file (virtual): ${path}`);
      return { type: 'Some', value: path };
    }

    return { type: 'None' };
  }

  /**
   * List files in directory
   */
  async listFiles(directory: string): Promise<OptionResult> {
    // Try real file system if in real mode
    if (this.mode === 'real' && this.directoryHandle) {
      try {
        const dirHandle = await this.getDirectoryHandle(directory);
        const files: string[] = [];

        for await (const entry of dirHandle.values()) {
          files.push(entry.name);
        }

        console.log(`[BrowserFS] Listed directory (real): ${directory} - ${files.length} entries`);
        return { type: 'Some', value: files };
      } catch (err) {
        console.warn(`[BrowserFS] Real FS list failed for ${directory}:`, err);
        return { type: 'None' };
      }
    }

    // Virtual mode
    const prefix = this.virtualPrefix + (directory ? directory + '/' : '');
    const files = new Set<string>();

    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key?.startsWith(prefix)) {
        const relativePath = key.substring(prefix.length);
        const fileName = relativePath.split('/')[0];
        if (fileName) {
          files.add(fileName);
        }
      }
    }

    const fileList = Array.from(files);
    console.log(`[BrowserFS] Listed directory (virtual): ${directory} - ${fileList.length} entries`);

    return fileList.length > 0
      ? { type: 'Some', value: fileList }
      : { type: 'None' };
  }

  /**
   * Helper: Get file handle from path
   */
  private async getFileHandle(
    path: string,
    options?: FileSystemGetFileOptions
  ): Promise<FileSystemFileHandle> {
    if (!this.directoryHandle) {
      throw new Error('No directory access');
    }

    const parts = path.split('/').filter(p => p);
    let currentDir = this.directoryHandle;

    // Navigate through directories
    for (let i = 0; i < parts.length - 1; i++) {
      currentDir = await currentDir.getDirectoryHandle(parts[i], { create: true });
    }

    // Get/create the file
    const fileName = parts[parts.length - 1];
    return await currentDir.getFileHandle(fileName, options);
  }

  /**
   * Helper: Get directory handle from path
   */
  private async getDirectoryHandle(path: string): Promise<FileSystemDirectoryHandle> {
    if (!this.directoryHandle) {
      throw new Error('No directory access');
    }

    if (!path || path === '.' || path === '/') {
      return this.directoryHandle;
    }

    const parts = path.split('/').filter(p => p);
    let currentDir = this.directoryHandle;

    for (const part of parts) {
      currentDir = await currentDir.getDirectoryHandle(part);
    }

    return currentDir;
  }
}

// Global singleton instance
export const browserFS = new BrowserFileSystem();

// Expose to WASM via window
(window as any).layerFileSystem = {
  // Storage mode control
  isSupported: () => browserFS.isFileSystemAccessSupported(),
  getMode: () => browserFS.getMode(),
  setMode: (mode: StorageMode) => browserFS.setMode(mode),
  requestAccess: () => browserFS.requestDirectoryAccess(),

  // File operations (all return promises that WASM will await)
  readFile: (path: string) => browserFS.readFile(path),
  writeFile: (path: string, content: string) => browserFS.writeFile(path, content),
  exists: (path: string) => browserFS.exists(path),
  deleteFile: (path: string) => browserFS.deleteFile(path),
  listFiles: (dir: string) => browserFS.listFiles(dir),

  // Synchronous versions for virtual mode only (for WASM deadlock fix)
  readFileSync: (path: string): OptionResult => {
    const key = browserFS['virtualPrefix'] + path;
    const content = localStorage.getItem(key);
    return content !== null
      ? { type: 'Some', value: content }
      : { type: 'None' };
  },

  writeFileSync: (path: string, content: string): OptionResult => {
    try {
      const key = browserFS['virtualPrefix'] + path;
      localStorage.setItem(key, content);
      console.log(`[BrowserFS] Wrote file (virtual): ${path}`);
      return { type: 'Some', value: content };
    } catch (err) {
      console.error(`[BrowserFS] localStorage write failed for ${path}:`, err);
      return { type: 'None' };
    }
  },

  existsSync: (path: string): boolean => {
    const key = browserFS['virtualPrefix'] + path;
    return localStorage.getItem(key) !== null;
  },

  deleteFileSync: (path: string): OptionResult => {
    const key = browserFS['virtualPrefix'] + path;
    const existed = localStorage.getItem(key) !== null;
    if (existed) {
      localStorage.removeItem(key);
      console.log(`[BrowserFS] Deleted file (virtual): ${path}`);
      return { type: 'Some', value: path };
    }
    return { type: 'None' };
  },

  listFilesSync: (dir: string): OptionResult => {
    const prefix = browserFS['virtualPrefix'] + (dir ? dir + '/' : '');
    const files: string[] = [];

    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && key.startsWith(prefix)) {
        const relativePath = key.substring(browserFS['virtualPrefix'].length);
        files.push(relativePath);
      }
    }

    return { type: 'Some', value: files };
  },
};

console.log('[BrowserFS] Initialized with mode:', browserFS.getMode());
console.log('[BrowserFS] File System Access API supported:', browserFS.isFileSystemAccessSupported());
