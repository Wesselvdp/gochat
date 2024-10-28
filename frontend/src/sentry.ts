// src/lib/sentry.js
import * as Sentry from "@sentry/browser";


export const initSentry = () => {
    console.log({env: import.meta.env})
    if (!import.meta.env.VITE_SENTRY_DSN) {
        console.warn('Sentry DSN not found');
        return;
    }


    Sentry.init({
        dsn: import.meta.env.VITE_SENTRY_DSN,
        integrations: [
            Sentry.browserTracingIntegration(),
            Sentry.globalHandlersIntegration({
                onerror: true,
                onunhandledrejection: true,
            })
        ],
        tracesSampleRate: 1.0,
        environment: import.meta.env.MODE,

        // Configure before sending the error
        // Capture all console errors in development
        beforeSend(event, hint) {
            const error = hint?.originalException;
            const isReferencedError = error instanceof ReferenceError;

            // Always capture ReferenceErrors
            if (isReferencedError) {
                return event;
            }

            // Development mode handling
            if (import.meta.env.DEV) {
                console.warn('Sentry error in development:', event);
                // Optionally still send certain types of errors in development
                return isReferencedError ? event : null;
            }

            return event;
        },
    });
};