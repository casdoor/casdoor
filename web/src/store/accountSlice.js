import {createSlice} from "@reduxjs/toolkit";

export const accountSlice = createSlice({
  name: "account",
  initialState: {
    value: null,
  },
  reducers: {
    setAccount: (state, action) => {
      state.value = action.payload;
    },
  },
});

export const {setAccount} = accountSlice.actions;

export const selectAccount = (state) => state.account.value;

export default accountSlice.reducer;
