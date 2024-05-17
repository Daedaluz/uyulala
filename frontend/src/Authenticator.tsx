import {useChallenge} from "./Hooks/useChallenge.ts";
import {CreateKey} from "./Components/CreateKey.tsx";
import {Sign} from "./Components/Sign.tsx";
import {ReactElement} from "react";
import {useSearchParams} from "react-router-dom";

function Authenticator() {
    const [params] = useSearchParams();
    const id = params.get("id");
    if (!id) {
        return <h3>Missing id</h3>
    }
    const {assertOptions, createOptions, error, loading, app, signData} = useChallenge(id);
    if (loading) {
        return <div>Loading...</div>
    }
    if (error) {
        return <h3>{error.msg}</h3>
    }
    let component: ReactElement
    if (createOptions) {
        component = <CreateKey id={id} challenge={createOptions}/>
    } else if (assertOptions) {
        component = <Sign id={id} challenge={assertOptions} app={app} signData={signData} />
    } else {
        component = <h3>Unknown challenge</h3>
    }
    return (
        {...component}
    )
}

export default Authenticator
