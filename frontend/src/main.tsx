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
import {createTheme, ThemeProvider} from "@mui/material";
import {AlertProvider} from "./Context/Alert.tsx";

const darkTheme = createTheme({
    palette: {
        mode: 'dark',
    }
})


ReactDOM.createRoot(document.getElementById('root')!).render(
    <ApiProvider>
        <ThemeProvider theme={darkTheme}>
            <AlertProvider>
                <BrowserRouter>
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
