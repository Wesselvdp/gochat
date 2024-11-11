import { test, expect, type Page } from '@playwright/test';

// Store credentials in environment variables for security
const MS_USERNAME = process.env.MS_USERNAME || "wessel@torgon.io";
const MS_PASSWORD = process.env.MS_PASSWORD || 'WaterGate7708';

// Helper function to handle Microsoft login
async function handleMicrosoftLogin(page: Page) {
  await page.goto('http://localhost:8080/login');
  await page.getByTestId('microsoft-login-button').click();

  await page.fill('input[type="email"]', MS_USERNAME);
  await page.getByRole('button', { name: 'Next' }).click();

  await page.waitForSelector('input[type="password"]', { state: 'visible', timeout: 10000 });
  await page.waitForTimeout(1000);
  await page.fill('input[type="password"]', MS_PASSWORD);
  await page.getByRole('button', { name: 'Sign in' }).click();

  await page.locator('#acceptButton').click();
}



test.describe('Authentication Flow', () => {
  test('should login with Microsoft account', async ({ page }) => {
    // // Handle the Microsoft login flow
    await handleMicrosoftLogin(page);

    await expect(page.locator('textarea[name="message"]')).toBeVisible();

    // Create a new chat
    await page.fill('#growingTextarea', "Say this is a test");
    // Wait for the API call to complete after pressing Enter
    await Promise.all([
      // Wait for the network request that matches your API endpoint
      page.waitForResponse(response =>
          response.url().includes('/send-message') &&
          response.status() === 200
      ),
      await page.keyboard.press('Enter')
    ]);

    await expect(page.locator('assistant-message .markdown-content')).toContainText("This is a test. ")
    // // Verify successful login (adjust according to your app's UI)
  });
});
