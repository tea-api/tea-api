import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';
import { marked } from 'marked';
import { Layout } from '@douyinfe/semi-ui';

const About = () => {
  const [about, setAbout] = useState('');
  const [aboutLoaded, setAboutLoaded] = useState(false);

  const displayAbout = async () => {
    setAbout(localStorage.getItem('about') || '');
    const res = await API.get('/api/about');
    const { success, message, data } = res.data;
    if (success) {
      let aboutContent = data;
      if (!data.startsWith('https://')) {
        aboutContent = marked.parse(data);
      }
      setAbout(aboutContent);
      localStorage.setItem('about', aboutContent);
    } else {
      showError(message);
      setAbout('加载关于内容失败...');
    }
    setAboutLoaded(true);
  };

  useEffect(() => {
    displayAbout().then();
  }, []);

  return (
    <>
      {aboutLoaded && about === '' ? (
        <>
          <Layout>
            <Layout.Header>
              <h3>关于</h3>
            </Layout.Header>
            <Layout.Content>
              <p>可在设置页面设置关于内容，支持 HTML & Markdown</p>
              Tea-API项目仓库地址：
              <a href='https://github.com/tea-api/tea-api'>
                https://github.com/tea-api/tea-api
              </a>
              <p>
                TeaAPI © 2023 CalciumIon | 基于 One API v0.5.4 © 2023
                JustSong。
              </p>
              <p>
                本项目根据MIT许可证授权，需在遵守Apache-2.0协议的前提下使用。
              </p>
            </Layout.Content>
          </Layout>
        </>
      ) : (
        <>
          {about.startsWith('https://') ? (
            <iframe
              src={about}
              style={{ width: '100%', height: '100vh', border: 'none' }}
            />
          ) : (
            <div
              style={{ fontSize: 'larger' }}
              dangerouslySetInnerHTML={{ __html: about }}
            ></div>
          )}
        </>
      )}
    </>
  );
};

export default About;
