import { API_BASE, APP_ENVIRONMENT, TELEMETRY_ENABLED } from '$lib/config';

interface ErrorContext {
	source?: string;
	url?: string;
}

interface ErrorDetails {
	message: string;
	name: string;
	stack: string;
}

const recentReports = new Map<string, number>();
const deduplicationWindowMs = 2_000;
let installed = false;

function errorDetails(value: unknown): ErrorDetails {
	if (value instanceof Error) {
		return {
			message: value.message || value.name,
			name: value.name,
			stack: value.stack ?? ''
		};
	}

	if (typeof value === 'string') {
		return { message: value, name: 'Error', stack: '' };
	}

	try {
		return { message: JSON.stringify(value), name: 'Error', stack: '' };
	} catch {
		return { message: String(value), name: 'Error', stack: '' };
	}
}

function shouldSend(details: ErrorDetails, source: string): boolean {
	const now = Date.now();
	const key = `${source}:${details.name}:${details.message}:${details.stack}`;
	const lastSentAt = recentReports.get(key);
	if (lastSentAt !== undefined && now - lastSentAt < deduplicationWindowMs) return false;

	recentReports.set(key, now);
	for (const [entry, sentAt] of recentReports) {
		if (now - sentAt >= deduplicationWindowMs) recentReports.delete(entry);
	}
	return true;
}

function sanitizeURL(value: string): string {
	try {
		const url = new URL(value, window.location.origin);
		url.username = '';
		url.password = '';
		url.search = '';
		url.hash = '';
		return url.toString();
	} catch {
		return '';
	}
}

export function reportError(value: unknown, context: ErrorContext = {}): void {
	if (!TELEMETRY_ENABLED || typeof window === 'undefined') return;

	const details = errorDetails(value);
	const source = context.source ?? 'application';
	if (!details.message || !shouldSend(details, source)) return;

	const body = JSON.stringify({
		...details,
		source,
		url: sanitizeURL(context.url ?? window.location.href),
		userAgent: navigator.userAgent,
		environment: APP_ENVIRONMENT
	});

	void fetch(new URL('/telemetry/errors', API_BASE), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body,
		keepalive: true
	}).catch(() => undefined);
}

export function installTelemetry(): () => void {
	if (!TELEMETRY_ENABLED || typeof window === 'undefined' || installed) return () => undefined;
	installed = true;

	const originalConsoleError = console.error;
	const handleWindowError = (event: ErrorEvent) => {
		reportError(event.error ?? event.message, { source: 'window.error', url: event.filename || undefined });
	};
	const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
		reportError(event.reason, { source: 'unhandledrejection' });
	};

	console.error = (...args: unknown[]) => {
		originalConsoleError(...args);
		const error = args.find((argument) => argument instanceof Error);
		reportError(error ?? args.map((argument) => errorDetails(argument).message).join(' '), {
			source: 'console.error'
		});
	};

	window.addEventListener('error', handleWindowError);
	window.addEventListener('unhandledrejection', handleUnhandledRejection);

	return () => {
		console.error = originalConsoleError;
		window.removeEventListener('error', handleWindowError);
		window.removeEventListener('unhandledrejection', handleUnhandledRejection);
		installed = false;
	};
}
