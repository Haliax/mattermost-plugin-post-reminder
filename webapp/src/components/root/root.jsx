import React from 'react';
import PropTypes from 'prop-types';

import {makeStyleFromTheme, changeOpacity} from 'mattermost-redux/utils/theme_utils';

import FullScreenModal from '../modals/full_screen_modal.jsx';

import './root.scss';

const PostUtils = window.PostUtils;

export default class Root extends React.Component {
    static propTypes = {
        visible: PropTypes.bool.isRequired,
        message: PropTypes.string.isRequired,
        postID: PropTypes.string.isRequired,
        close: PropTypes.func.isRequired,
        submit: PropTypes.func.isRequired,
        theme: PropTypes.object.isRequired,
    }
    constructor(props) {
        super(props);

        this.state = {
            message: null,
            sendTo: null,
            attachToThread: false,
            reminderType: 'dm',
            reminderDate: '3600',
            durationNumber: 1,
            durationType: 'hours',
            previewMarkdown: false,
        };
    }

    static getDerivedStateFromProps(props, state) {
        if (props.visible && state.message == null) {
            return {message: props.message};
        }
        if (!props.visible && (state.message != null || state.sendTo != null)) {
            return {message: null, sendTo: null, attachToThread: false, previewMarkdown: false};
        }
        return null;
    }

    handleAttachChange = (e) => {
        const value = e.target.checked;
        if (value !== this.state.attachToThread) {
            this.setState({
                attachToThread: value,
            });
        }
    }

    handleTypeChange = (type) => {
        this.setState({reminderType: type});
    }

    submit = () => {
        const {submit, close, postID} = this.props;
        const {message, sendTo, reminderType, reminderDate} = this.state;
        submit(message, sendTo, postID, reminderType, reminderDate);
        close();
    }

    render() {
        const {visible, theme, close} = this.props;

        if (!visible) {
            return null;
        }

        const {message} = this.state;

        const style = getStyle(theme);
        const activeClass = 'btn btn-primary';
        const inactiveClass = 'btn';
        const writeButtonClass = this.state.previewMarkdown ? inactiveClass : activeClass;
        const previewButtonClass = this.state.previewMarkdown ? activeClass : inactiveClass;
        const unreadButtonClass = this.state.reminderType === 'unread' ? activeClass : inactiveClass;
        const dMButtonClass = this.state.reminderType === 'dm' ? activeClass : inactiveClass;
        const messageWrapper = this.state.reminderType === 'dm' ? '' : 'hidden';

        return (
            <FullScreenModal
                show={visible}
                onClose={close}
            >
                <div
                    style={style.modal}
                    className='PostReminderPluginRootModal'
                >
                    <h1>{'Add a Reminder'}</h1>
                    <div className='postreminderplugin-issue'>
                        <div className='hidden'>
                            <h3>
                                {'Reminder Type'}
                            </h3>
                            <div className='btn-group'>
                                <button
                                    className={unreadButtonClass}
                                    onClick={() => {
                                        this.handleTypeChange('unread');
                                    }}
                                >
                                    {'Mark the message as unread!'}
                                </button>
                                <button
                                    className={dMButtonClass}
                                    onClick={() => {
                                        this.handleTypeChange('dm');
                                    }}
                                >
                                    {'Send me a DM!'}
                                </button>
                            </div>
                            <div className={messageWrapper}>
                                <h3>
                                    {'Reminder Message'}
                                </h3>
                                <div className='btn-group'>
                                    <button
                                        className={writeButtonClass}
                                        onClick={() => {
                                            this.setState({previewMarkdown: false});
                                        }}
                                    >
                                        {'Write'}
                                    </button>
                                    <button
                                        className={previewButtonClass}
                                        onClick={() => {
                                            this.setState({previewMarkdown: true});
                                        }}
                                    >
                                        {'Preview'}
                                    </button>
                                </div>
                                {this.state.previewMarkdown ? (
                                    <div
                                        className='postreminderplugin-input'
                                        style={style.markdown}
                                    >
                                        {PostUtils.messageHtmlToComponent(
                                            PostUtils.formatText(this.state.message),
                                        )}
                                    </div>
                                ) : (
                                    <textarea
                                        className='postreminderplugin-input'
                                        style={style.textarea}
                                        value={message}
                                        onChange={(e) => this.setState({message: e.target.value})}
                                    />)
                                }
                            </div>
                        </div>
                        <h3>
                            {'When do you want to be reminded?'}
                        </h3>
                        <div className='postreminderplugin-duration-container'>
                            <input
                                type={'number'}
                                value={this.state.durationNumber}
                                onChange={(e) => {
                                    this.setState({durationNumber: ((e.target.value ? parseInt(e.target.value, 10) : 0))});
                                    this.setState({reminderDate: ((e.target.value ? parseInt(e.target.value, 10) : 0) * 60 * 1000)});
                                }}
                            />
                            <b>{' Minutes'}</b>
                        </div>
                    </div>
                    <div className='postreminderplugin-button-container'>
                        <button
                            className={'btn btn-primary'}
                            style={message ? style.button : style.inactiveButton}
                            onClick={this.submit}
                            disabled={!message}
                        >
                            {'Add Reminder'}
                        </button>
                    </div>
                    <div className='postreminderplugin-divider'/>
                    <div className='postreminderplugin-clarification'>
                        <div className='postreminderplugin-question'>
                            {'What does this do?'}
                        </div>
                        <div className='postreminderplugin-answer'>
                            {'Adding a Reminder will mark the message as unread after a certain amount of time or send you a reminder via DM.'}
                        </div>
                    </div>
                </div>
            </FullScreenModal>
        );
    }
}

const getStyle = makeStyleFromTheme((theme) => {
    return {
        modal: {
            color: changeOpacity(theme.centerChannelColor, 0.88),
        },
        textarea: {
            backgroundColor: theme.centerChannelBg,
        },
        helpText: {
            color: changeOpacity(theme.centerChannelColor, 0.64),
        },
        button: {
            color: theme.buttonColor,
            backgroundColor: theme.buttonBg,
        },
        inactiveButton: {
            color: changeOpacity(theme.buttonColor, 0.88),
            backgroundColor: changeOpacity(theme.buttonBg, 0.32),
        },
        markdown: {
            minHeight: '149px',
            fontSize: '16px',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'end',
        },
    };
});
