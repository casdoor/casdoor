import {combineReducers, configureStore} from "@reduxjs/toolkit";
import accountReducer from "./accountSlice";
import themeReducer from "./themeSlice";

const rootReducer = combineReducers({
  account: accountReducer,
  theme: themeReducer,
});

export default configureStore({
  reducer: rootReducer,
});
