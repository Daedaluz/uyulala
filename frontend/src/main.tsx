import ReactDOM from 'react-dom/client'
import './index.css'
import '@fontsource/roboto/300.css'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import {BrowserRouter, Route, Routes} from "react-router-dom";
import {RootLayout} from "./Layout/RootLayout.tsx";
import Authenticator from "./Authenticator.tsx";
import {Authorize} from "./Authorize.tsx";
import {Home} from "./Home.tsx";
import {DemoPage} from "./Demo/Demo.tsx";
import {ApiProvider} from "./Context/Api.tsx";
import {colors, createTheme, ThemeProvider} from "@mui/material";
import {AlertProvider} from "./Context/Alert.tsx";

const darkTheme = createTheme({
    palette: {
        mode: 'light',
        text: {
            primary: "#ffffff",
            secondary: "#FFFFFF",
            disabled: "#AAAAAA"
        },
        common: {
            white: "#FFFFFF",
            black: "#000000"
        },
        background: {
            default: "#FFFFFF",
            paper: "#a0a0a0ee"
        },
        primary: colors.blue,

//        background: {
//            default: '#282828',
//            paper: '#292929'
//        },
//        text: {
//            primary: '#a0a0a0',
//            disabled: '#404040',
//            secondary: '#a2a2a2',
//        },
//        primary: colors.grey,
//        secondary: colors.grey,
//        common: {
//            black: '#252525',
//            white: '#505050'
//        },
//        grey: colors.grey,
//        info: colors.grey,
//        error: colors.brown,
//        warning: colors.amber
    }
})


ReactDOM.createRoot(document.getElementById('root')!).render(
    <ApiProvider>
        <ThemeProvider theme={darkTheme}>
            <AlertProvider>
                <BrowserRouter future={{
                    v7_relativeSplatPath: true,
                    v7_startTransition: true,
                }}>
                    <Routes>
                        <Route path="/" element={<RootLayout/>}>
                            <Route index element={<Home/>}/>
                            <Route path="/authenticator" element={<Authenticator/>}/>
                            <Route path="/authorize" element={<Authorize/>}/>
                            <Route path="/demo" element={<DemoPage/>}/>
                        </Route>
                    </Routes>
                </BrowserRouter>
            </AlertProvider>
        </ThemeProvider>
    </ApiProvider>
)
