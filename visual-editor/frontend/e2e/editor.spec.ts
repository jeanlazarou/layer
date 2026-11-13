import { expect, test, type Page } from '@playwright/test'

// Each test starts from a fresh browser context, so the editor shows the
// default project (debug wrapping core.factorial) and localStorage is empty.

/** Open the editor and wait for the interpreter-backed UI and Monaco. */
async function openEditor(page: Page) {
  await page.goto('./')
  await expect(page.getByRole('button', { name: 'Run', exact: true })).toBeVisible()
  await page.waitForFunction(() => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const monaco = (window as any).monaco
    return !!monaco && monaco.editor.getEditors().length > 0
  })
}

/** Replace the code of the buffer currently shown in the editor. */
async function setCode(page: Page, code: string) {
  await page.evaluate(text => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const monaco = (window as any).monaco
    monaco.editor.getEditors()[0].getModel().setValue(text)
  }, code)
}

function markers(page: Page) {
  return page.evaluate(() =>
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).monaco.editor.getModelMarkers({}).map((m: { message: string }) => m.message),
  )
}

function layerCard(page: Page, name: string) {
  return page.locator('.MuiCard-root').filter({ has: page.locator('p', { hasText: new RegExp(`^${name}$`) }) })
}

async function run(page: Page) {
  await page.getByRole('button', { name: 'Run', exact: true }).click()
}

async function openExample(page: Page, title: RegExp) {
  await page.getByRole('button', { name: 'Examples' }).click()
  await page.getByRole('menuitem', { name: title }).click()
  await expect(page.getByText(/^Opened /)).toBeVisible()
}

const output = (page: Page) => page.locator('body')

const CORE_FACTORIAL = 'fn factorial(n) =\n  match n with\n  | 0 -> 1\n  | n -> n * factorial(n - 1)'

test('running the default project shows the debug layer logs and the result', async ({ page }) => {
  await openEditor(page)
  await run(page)
  await expect(output(page)).toContainText('⇒ 120')
  await expect(output(page)).toContainText('[INFO] factorial(5)')
  await expect(output(page)).toContainText('[INFO] factorial(0)')
})

test('hiding a layer re-runs the program without it', async ({ page }) => {
  await openEditor(page)
  await run(page)
  await expect(output(page)).toContainText('[INFO] factorial(5)')

  await layerCard(page, 'debug').locator('button[title="Disable layer"]').click()
  await expect(output(page)).not.toContainText('[INFO] factorial(5)')
  await expect(output(page)).toContainText('⇒ 120')

  await layerCard(page, 'debug').locator('button[title="Enable layer"]').click()
  await expect(output(page)).toContainText('[INFO] factorial(5)')
})

test('the resolution panel shows which layer answers each call', async ({ page }) => {
  await openEditor(page)
  const panel = page.getByText('Who answers each call?').locator('..')
  await expect(panel).toContainText('factorial()')
  await expect(panel).toContainText('app()')
})

test('typing a mistake shows a live diagnostic before running', async ({ page }) => {
  await openEditor(page)
  await layerCard(page, 'core').click()
  await setCode(page, CORE_FACTORIAL.replace('factorial(n - 1)', 'factorail(n - 1)'))

  await expect.poll(() => markers(page)).toContainEqual(expect.stringContaining("undefined name 'factorail'"))
  await expect(page.getByText('1 error', { exact: true })).toBeVisible()
})

test('runtime errors point at the layer and line, and clicking one jumps there', async ({ page }) => {
  await openEditor(page)
  await layerCard(page, 'core').click()
  await setCode(page, CORE_FACTORIAL.replace('factorial(n - 1)', 'factorial(n - 1) + "x"'))
  await expect.poll(() => markers(page)).toEqual([])

  await run(page)
  const error = page.getByText(/runtime error/).locator('..')
  await expect(error).toContainText('at core:4:')
  await expect(error).toContainText('operator + needs two numbers or two strings, got integer and string')

  // Jump from another buffer back to the failing layer.
  await page.getByText('Main Expression').click()
  await error.click()
  await expect(page.getByText('pure layer')).toBeVisible()
})

test('UI results render as live widgets that run Layer code', async ({ page }) => {
  await openEditor(page)
  await page.getByText('Main Expression').click()
  await setCode(page, 'app()')
  await run(page)

  await page.getByPlaceholder('Enter a number').fill('6')
  await page.getByRole('button', { name: 'Compute' }).click()
  await expect(output(page)).toContainText('720')
})

test('the fallback example loads its blend mode and impure layer, and runs', async ({ page }) => {
  await openEditor(page)
  await openExample(page, /Blend mode: fallback/)
  await expect(page.getByRole('combobox')).toHaveText('blend: fallback')
  await expect(layerCard(page, 'cache')).toContainText('impure')

  await run(page)
  await expect(output(page)).toContainText('⇒ ["cached-hello", "computed-value-for-something-new"]')
  await expect(page.getByText('fetch()')).toBeVisible()
})

test('calculator buttons write into the display input (target)', async ({ page }) => {
  await openEditor(page)
  await openExample(page, /Interactive Calculator/)
  await run(page)

  for (const key of ['7', '×', '6', '=']) {
    await page.getByRole('button', { name: key, exact: true }).click()
  }
  await expect(page.getByPlaceholder('0', { exact: true })).toHaveValue('42')
})

test('the workspace survives a reload', async ({ page }) => {
  await openEditor(page)
  await openExample(page, /Interactive Calculator/)
  await expect(page.getByRole('banner')).toContainText('calculator')

  // Autosave is debounced; give it a moment before reloading.
  await page.waitForTimeout(600)
  await page.reload()
  await expect(page.getByRole('banner')).toContainText('calculator')
  await expect(layerCard(page, 'ui')).toBeVisible()
})

test('the help dialog opens', async ({ page }) => {
  await openEditor(page)
  await page.getByRole('button', { name: 'Help' }).click()
  await expect(page.getByRole('dialog')).toContainText('Layer in two minutes')
})
