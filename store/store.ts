import { configureStore } from '@reduxjs/toolkit'
import zonesReducer from './zonesSlice'

export const store = configureStore({
  reducer: {
    zones: zonesReducer,
  },
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch