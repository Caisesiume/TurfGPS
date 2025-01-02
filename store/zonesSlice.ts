import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { Zone } from '../lib/types/zone'

interface ZonesState {
  items: Record<string, Zone>
  loading: boolean
  error: string | null
}

const initialState: ZonesState = {
  items: {},
  loading: false,
  error: null,
}

const zonesSlice = createSlice({
  name: 'zones',
  initialState,
  reducers: {
    setZonesLoading: (state) => {
      state.loading = true
      state.error = null
    },
    setZonesSuccess: (state, action: PayloadAction<Zone[]>) => {
      state.loading = false
      action.payload.forEach(zone => {
        state.items[`${zone.latitude},${zone.longitude}`] = zone
      })
    },
    setZonesError: (state, action: PayloadAction<string>) => {
      state.loading = false
      state.error = action.payload
    },
  },
})

export const { setZonesLoading, setZonesSuccess, setZonesError } = zonesSlice.actions
export default zonesSlice.reducer