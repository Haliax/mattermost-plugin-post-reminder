import {Client4} from 'mattermost-redux/client';

import {
    OPEN_ROOT_MODAL,
    CLOSE_ROOT_MODAL,
} from './action_types';

import {getPluginServerRoute} from './selectors';

export const openRootModal = (postID) => (dispatch) => {
    dispatch({
        type: OPEN_ROOT_MODAL,
        postID,
    });
};

export const closeRootModal = () => (dispatch) => {
    dispatch({
        type: CLOSE_ROOT_MODAL,
    });
};

export const add = (message, sendTo, postID, reminderType, reminderDate) => async (dispatch, getState) => {
    await fetch(getPluginServerRoute(getState()) + '/add', Client4.getOptions({
        method: 'post',
        body: JSON.stringify({message, send_to: sendTo, post_id: postID, remember_at: reminderDate.toString(), reminder_type: reminderType}),
    }));
};
