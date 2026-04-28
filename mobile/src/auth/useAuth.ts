import { useAuthStore } from './store';

export const useAuth = () => useAuthStore();
export const useAuthStatus = () => useAuthStore((s) => s.status);
export const useUser = () => useAuthStore((s) => s.user);
