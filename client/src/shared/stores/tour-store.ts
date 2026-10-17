import { create } from 'zustand'

interface TourState {
  /** Grows by one each time someone asks for the guided tour; the tour starts on every change. */
  startRequests: number
  requestStart: () => void
}

/** Connects the account menu to the guided tour without the two importing each other. */
export const useTourStore = create<TourState>()((set) => ({
  startRequests: 0,
  requestStart: () => set((state) => ({ startRequests: state.startRequests + 1 })),
}))
