import {id as pluginId} from './manifest';

import Root from './components/root';

import {openRootModal} from './actions';
import reducer from './reducer';

let activityFunc;

export default class Plugin {
    initialize(registry, store) {
        registry.registerReducer(reducer);
        registry.registerRootComponent(Root);

        registry.registerPostDropdownMenuAction(
            'Add Reminder',
            (postID) => {
                store.dispatch(openRootModal(postID));
            },
        );

        document.addEventListener('click', activityFunc);
    }

    deinitialize() {
        document.removeEventListener('click', activityFunc);
    }
}

window.registerPlugin(pluginId, new Plugin());
