// src/lib/sentry.js
import * as Sentry from "@sentry/browser";


export const initSentry = () => {

    if (import.meta.env.MODE === 'development') {
        return console.log('Sentry not initialised in dev mode')

    }
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


            return event;
        },
    });
};