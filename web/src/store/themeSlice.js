import {createSlice} from "@reduxjs/toolkit";
import * as Conf from "../Conf";
import * as Setting from "../Setting";

function getLogo(themes) {
  if (themes.includes("dark")) {
    return `${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256_dark.png`;
  } else {
    return `${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256.png`;
  }
}

export const themeSlice = createSlice({
  name: "theme",
  initialState: {
    value: Conf.ThemeDefault,
    logo: getLogo(Setting.getAlgorithmNames(Conf.ThemeDefault)),
    themeAlgorithm: ["default"],
  },
  reducers: {
    setTheme: (state, action) => {
      state.value = action.payload;
    },
    setLogo(state, action) {
      if (action.payload instanceof Array) {
        state.logo = getLogo(action.payload);
        return;
      }
      state.logo = getLogo(Setting.getAlgorithmNames(action.payload));
    },
    setThemeAlgorithm(state, action) {
      if (action.payload instanceof Array) {
        state.themeAlgorithm = action.payload.sort().reverse();
        return;
      }
      state.themeAlgorithm = Setting.getAlgorithmNames(action.payload);
    },
    setThemeAndUpdate(state, action) {
      state.value = action.payload?.theme;
      state.logo = getLogo(Setting.getAlgorithmNames(action.payload));
      state.themeAlgorithm = Setting.getAlgorithmNames(action.payload);
    },
  },
});

export const {
  setTheme,
  setLogo,
  setThemeAlgorithm,
  setThemeAndUpdate,
} = themeSlice.actions;

export const selectTheme = (state) => state.theme.value;
export const selectLogo = (state) => state.theme.logo;
export const selectThemeAlgorithm = (state) => state.theme.themeAlgorithm;

export default themeSlice.reducer;
