import {getConfig} from 'mattermost-redux/selectors/entities/general';

import {id as pluginId} from './manifest';

const getPluginState = (state) => state['plugins-' + pluginId] || {};

export const isRootModalVisible = (state) => getPluginState(state).rootModalVisible;
export const getPostID = (state) => getPluginState(state).postID;
export const getMessage = (state) => {
    const postID = getPluginState(state).postID;
    if (!postID) {
        return '';
    }
    const post = state.entities.posts.posts[postID];
    if (!post) {
        return '';
    }
    return post.message;
};

// TODO: Move this into mattermost-redux or mattermost-webapp.
export const getSiteURL = (state) => {
    const config = getConfig(state);

    let basePath = '';
    if (config && config.SiteURL) {
        basePath = new URL(config.SiteURL).pathname;

        if (basePath && basePath[basePath.length - 1] === '/') {
            basePath = basePath.substr(0, basePath.length - 1);
        }
    }

    return basePath;
};

export const getPluginServerRoute = (state) => {
    const siteURL = getSiteURL(state);
    return siteURL + '/plugins/' + pluginId;
};
