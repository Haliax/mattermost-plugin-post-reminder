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
            attachToThread: false,
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
        if (!props.visible && (state.message != null)) {
            return {message: null, attachToThread: false, previewMarkdown: false};
        }
        return null;
    }

    setStateAsync(state) {
        return new Promise((resolve) => {
            this.setState(state, resolve);
        });
    }

    async calculateDuration() {
        let multiplier;
        switch (this.state.durationType) {
        case 'minutes':
            multiplier = 60;
            break;
        case 'hours':
            multiplier = 3600;
            break;
        case 'days':
            multiplier = 86400;
            break;
        default:
            multiplier = 3600;
            break;
        }

        await this.setStateAsync({reminderDate: this.state.durationNumber * multiplier * 1000});
    }

    submit = async () => {
        await this.calculateDuration();
        const {submit, close, postID} = this.props;
        const {message, reminderDate} = this.state;
        submit(message, postID, reminderDate);
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
                        <h3>
                            {'When do you want to be reminded?'}
                        </h3>
                        <div className='postreminderplugin-duration-container'>
                            <input
                                type={'number'}
                                value={this.state.durationNumber}
                                onChange={async (e) => {
                                    await this.setStateAsync({durationNumber: ((e.target.value ? parseInt(e.target.value, 10) : 0))});
                                }}
                            />
                            <select
                                value={this.state.durationType}
                                onChange={async (e) => {
                                    await this.setStateAsync({durationType: e.target.value});
                                }}
                            >
                                <option value={'minutes'}>{'Minutes'}</option>
                                <option value={'hours'}>{'Hours'}</option>
                                <option value={'days'}>{'Days'}</option>
                            </select>
                        </div>

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
                            {'Adding a Reminder will send you a reminder via DM after a certain amount of time.'}
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
