import type enCommon from '../locales/en-US/common.json'
import type enAuth from '../locales/en-US/auth.json'
import type enLobby from '../locales/en-US/lobby.json'
import type enTable from '../locales/en-US/table.json'
import type enErrors from '../locales/en-US/errors.json'

// Augment i18next with the en-US canonical types so `t('unknown.key')` fails at compile time.
declare module 'i18next' {
  interface CustomTypeOptions {
    defaultNS: 'common'
    resources: {
      common: typeof enCommon
      auth: typeof enAuth
      lobby: typeof enLobby
      table: typeof enTable
      errors: typeof enErrors
    }
  }
}
