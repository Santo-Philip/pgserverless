import { env } from '$env/dynamic/public';

export const APP_NAME = env.PUBLIC_APP_NAME || 'PgServerless Admin';
export const APP_NAME_SHORT = env.PUBLIC_APP_NAME_SHORT || 'PgAdmin';
export const APP_LOGO_LETTER = APP_NAME_SHORT.charAt(0).toUpperCase();
