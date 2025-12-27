// Layer WASM Module Loader and Interface
import './browserFileSystem';

interface ExecuteResult {
  success: boolean;
  result?: string;
  logs?: string[];
  error?: string;
}

interface ParseResult {
  success: boolean;
  enabledLayers?: string[];
  layers?: Record<string, string>;
  mainExpression?: string;
  error?: string;
}

declare global {
  interface Window {
    Go: any;
    executeLayer: (code: string) => ExecuteResult;
    parseLayerFile: (content: string) => ParseResult;
  }
}

class LayerWasm {
  private loaded: boolean = false;
  private loading: Promise<void> | null = null;

  /**
   * Load the WASM module
   */
  async load(): Promise<void> {
    if (this.loaded) {
      return;
    }

    if (this.loading) {
      return this.loading;
    }

    this.loading = this.loadWasm();
    await this.loading;
    this.loaded = true;
  }

  private async loadWasm(): Promise<void> {
    try {
      // Load the wasm_exec.js script
      await this.loadScript('/wasm_exec.js');

      // Create Go instance
      const go = new window.Go();

      // Fetch and instantiate the WASM module
      const response = await fetch('/layer.wasm');
      const buffer = await response.arrayBuffer();
      const result = await WebAssembly.instantiate(buffer, go.importObject);

      // Run the WASM module
      go.run(result.instance);

      console.log('Layer WASM module loaded successfully!');
    } catch (error) {
      console.error('Failed to load Layer WASM module:', error);
      throw new Error(`WASM loading failed: ${error}`);
    }
  }

  private loadScript(src: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const script = document.createElement('script');
      script.src = src;
      script.onload = () => resolve();
      script.onerror = () => reject(new Error(`Failed to load script: ${src}`));
      document.head.appendChild(script);
    });
  }

  /**
   * Execute Layer source code
   */
  async execute(code: string): Promise<ExecuteResult> {
    await this.load();

    if (!window.executeLayer) {
      throw new Error('executeLayer function not available');
    }

    try {
      const result = window.executeLayer(code);
      return result;
    } catch (error) {
      return {
        success: false,
        error: `Execution error: ${error}`,
      };
    }
  }

  /**
   * Parse a .layer file
   */
  async parseFile(content: string): Promise<ParseResult> {
    await this.load();

    if (!window.parseLayerFile) {
      throw new Error('parseLayerFile function not available');
    }

    try {
      const result = window.parseLayerFile(content);
      return result;
    } catch (error) {
      return {
        success: false,
        error: `Parse error: ${error}`,
      };
    }
  }

  /**
   * Check if WASM module is loaded
   */
  isLoaded(): boolean {
    return this.loaded;
  }
}

// Create singleton instance
const layerWasm = new LayerWasm();

export default layerWasm;
export type { ExecuteResult, ParseResult };
