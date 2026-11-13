// Monaco language definition for Layer: syntax highlighting (including
// "{expr}" interpolation in double-quoted strings), brackets/comments, and
// keyword + built-in completions.
import type { Monaco } from '@monaco-editor/react'

export const LAYER_LANGUAGE_ID = 'layer'

const KEYWORDS = ['fn', 'let', 'in', 'match', 'with', 'when', 'if', 'then', 'else', 'and', 'or', 'not', 'blend']

const BUILTINS: Record<string, string[]> = {
  core: [
    'map', 'filter', 'reduce', 'fold', 'head', 'tail', 'length', 'concat', 'take', 'drop',
    'split', 'str_concat', 'str_length', 'join', 'to_string', 'to_float', 'to_int',
    'id', 'const', 'compose', 'print',
  ],
  debug: ['log', 'trace', 'inspect', 'assert_equal', 'benchmark', 'breakpoint', 'get_logs', 'clear_logs'],
  ui: ['view', 'text', 'button', 'input', 'container', 'row', 'column', 'render'],
  data: [
    'Some', 'None', 'is_some', 'is_none', 'unwrap', 'map_option', 'flatmap_option',
    'read_file', 'write_file', 'exists', 'delete', 'list_files', 'read_json', 'write_json',
    'parse_json', 'to_json', 'get', 'set', 'delete_key', 'keys', 'clear',
  ],
}

let registered = false

export function registerLayerLanguage(monaco: Monaco): void {
  if (registered) return
  registered = true

  monaco.languages.register({ id: LAYER_LANGUAGE_ID, extensions: ['.layer'] })

  monaco.languages.setLanguageConfiguration(LAYER_LANGUAGE_ID, {
    comments: { lineComment: '#' },
    brackets: [['{', '}'], ['[', ']'], ['(', ')']],
    autoClosingPairs: [
      { open: '{', close: '}' },
      { open: '[', close: ']' },
      { open: '(', close: ')' },
      { open: '"', close: '"', notIn: ['string'] },
      { open: "'", close: "'", notIn: ['string'] },
    ],
    surroundingPairs: [
      { open: '{', close: '}' },
      { open: '[', close: ']' },
      { open: '(', close: ')' },
      { open: '"', close: '"' },
    ],
  })

  monaco.languages.setMonarchTokensProvider(LAYER_LANGUAGE_ID, {
    keywords: [...KEYWORDS, 'true', 'false'],
    builtins: Object.values(BUILTINS).flat(),
    tokenizer: {
      root: [
        [/#.*$/, 'comment'],
        [/@(layers|layer|which|inspect|diff)\b/, 'keyword.directive'],
        [/<(pure|impure)>/, 'annotation'],
        [/λ/, 'keyword'],
        // layer-qualified access: core.factorial
        [/[a-zA-Z_]\w*(?=\.[a-zA-Z_])/, 'type.identifier'],
        [/[a-zA-Z_]\w*/, { cases: { '@keywords': 'keyword', '@builtins': 'predefined', '@default': 'identifier' } }],
        [/\d+\.\d+/, 'number.float'],
        [/\d+/, 'number'],
        [/"/, 'string', '@dstring'],
        [/'/, 'string', '@sstring'],
        [/->|==|!=|<=|>=|\.\.\.|[+\-*/%^<>=|?]/, 'operator'],
        [/[{}()[\]]/, '@brackets'],
      ],
      dstring: [
        [/\\./, 'string.escape'],
        [/\{(?=[^\s"'\\}])/, { token: 'delimiter.interpolation', next: '@interp' }],
        [/[^"\\{]+/, 'string'],
        [/\{/, 'string'],
        [/"/, 'string', '@pop'],
      ],
      interp: [
        [/\}/, { token: 'delimiter.interpolation', next: '@pop' }],
        [/'[^']*'/, 'string'],
        [/[a-zA-Z_]\w*/, { cases: { '@keywords': 'keyword', '@builtins': 'predefined', '@default': 'identifier' } }],
        [/\d+(\.\d+)?/, 'number'],
        [/[^}'a-zA-Z_\d]+/, 'operator'],
      ],
      sstring: [
        [/\\./, 'string.escape'],
        [/[^'\\]+/, 'string'],
        [/'/, 'string', '@pop'],
      ],
    },
  })

  monaco.languages.registerCompletionItemProvider(LAYER_LANGUAGE_ID, {
    provideCompletionItems(model, position) {
      const word = model.getWordUntilPosition(position)
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn,
      }
      const kinds = monaco.languages.CompletionItemKind
      const snippet = monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
      const suggestions = [
        ...Object.entries(BUILTINS).flatMap(([layer, names]) =>
          names.map(name => ({
            label: name,
            kind: kinds.Function,
            detail: `${layer} layer`,
            insertText: name,
            range,
          })),
        ),
        ...KEYWORDS.map(k => ({ label: k, kind: kinds.Keyword, insertText: k, range })),
        {
          label: 'fn',
          kind: kinds.Snippet,
          detail: 'function definition',
          insertText: 'fn ${1:name}(${2:x}) =\n  ${3:x}',
          insertTextRules: snippet,
          range,
        },
        {
          label: 'match',
          kind: kinds.Snippet,
          detail: 'pattern match',
          insertText: 'match ${1:value} with\n| ${2:0} -> ${3:result}\n| ${4:_} -> ${5:other}',
          insertTextRules: snippet,
          range,
        },
        {
          label: 'let',
          kind: kinds.Snippet,
          detail: 'let binding',
          insertText: 'let ${1:name} = ${2:value} in\n  ${3:name}',
          insertTextRules: snippet,
          range,
        },
        {
          label: 'if',
          kind: kinds.Snippet,
          detail: 'conditional',
          insertText: 'if ${1:condition} then ${2:a} else ${3:b}',
          insertTextRules: snippet,
          range,
        },
      ]
      return { suggestions }
    },
  })

  monaco.editor.defineTheme('layer-light', {
    base: 'vs',
    inherit: true,
    rules: [
      { token: 'keyword.directive', foreground: '7b1fa2', fontStyle: 'bold' },
      { token: 'annotation', foreground: 'e65100' },
      { token: 'type.identifier', foreground: '00796b', fontStyle: 'bold' },
      { token: 'predefined', foreground: '1565c0' },
      { token: 'delimiter.interpolation', foreground: 'e65100', fontStyle: 'bold' },
    ],
    colors: {},
  })
}
