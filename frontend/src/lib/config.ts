const runtimeOrigin = typeof window !== 'undefined' ? window.location.origin : 'http://localhost';
const useLocalDevelopmentServices = import.meta.env.DEV || import.meta.env.VITE_APP_TARGET === 'desktop';

export const API_BASE =
	import.meta.env.VITE_API_BASE_URL || (useLocalDevelopmentServices ? 'http://localhost:8088' : runtimeOrigin);

export const ZIG_API_BASE =
	import.meta.env.VITE_ZIG_API_BASE_URL || (useLocalDevelopmentServices ? 'http://localhost:8089' : runtimeOrigin);

export const TELEMETRY_ENABLED = import.meta.env.VITE_TELEMETRY_ENABLED === 'true';
export const APP_ENVIRONMENT = import.meta.env.VITE_APP_ENV || import.meta.env.MODE;
