# Expo + Effect pattern

Keep React Native rendering framework-native and put fallible async work in
Effect programs. Components consume derived hooks from
`mobile/src/auth/useAuth.ts`; the Atom model lives in
`mobile/src/auth/store.ts` <!-- expect: authStateAtom -->.

Import narrow Effect Atom entrypoints. Do not import the package barrel: it
re-exports RPC modules that are unnecessary in Expo.

```ts
import * as Atom from "@effect-atom/atom/Atom"
import { useAtomValue } from "@effect-atom/atom-react/Hooks"
import { RegistryProvider } from "@effect-atom/atom-react/RegistryContext"
```

Wrap SecureStore and fetch failures with `Effect.tryPromise`. Decode API JSON
using schemas from `@initium/core`. Run Effects only through action atoms or at
a framework/test boundary.
