import { useAtomSet, useAtomValue } from '@effect-atom/atom-react/Hooks';
import { useCallback, useMemo } from 'react';

import {
  authError,
  authStateAtom,
  authStatus,
  bootstrapAuthAtom,
  logoutAtom,
  requestMagicLinkAtom,
  verifyGoogleAtom,
  verifyMagicLinkAtom,
} from './store';

export const useAuth = () => {
  const state = useAtomValue(authStateAtom);
  const runBootstrap = useAtomSet(bootstrapAuthAtom, { mode: 'promise' });
  const runRequestMagicLink = useAtomSet(requestMagicLinkAtom, { mode: 'promise' });
  const runVerifyMagicLink = useAtomSet(verifyMagicLinkAtom, { mode: 'promise' });
  const runVerifyGoogle = useAtomSet(verifyGoogleAtom, { mode: 'promise' });
  const runLogout = useAtomSet(logoutAtom, { mode: 'promise' });

  const bootstrap = useCallback(async () => {
    await runBootstrap(undefined);
  }, [runBootstrap]);
  const requestMagicLink = useCallback(
    async (email: string) => {
      await runRequestMagicLink(email);
    },
    [runRequestMagicLink],
  );
  const verifyMagicLink = useCallback(
    async (token: string) => {
      await runVerifyMagicLink(token);
    },
    [runVerifyMagicLink],
  );
  const verifyGoogle = useCallback(
    async (idToken: string) => {
      await runVerifyGoogle(idToken);
    },
    [runVerifyGoogle],
  );
  const logout = useCallback(async () => {
    await runLogout(undefined);
  }, [runLogout]);

  return useMemo(
    () => ({
      status: authStatus(state),
      user: state._tag === 'Authenticated' ? state.user : null,
      error: authError(state),
      bootstrap,
      requestMagicLink,
      verifyMagicLink,
      verifyGoogle,
      logout,
    }),
    [bootstrap, logout, requestMagicLink, state, verifyGoogle, verifyMagicLink],
  );
};

export const useAuthStatus = () => authStatus(useAtomValue(authStateAtom));

export const useUser = () => {
  const state = useAtomValue(authStateAtom);
  return state._tag === 'Authenticated' ? state.user : null;
};
