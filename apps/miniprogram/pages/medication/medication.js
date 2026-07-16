Page({
  data: {
    medications: [
      { id: 1, name: '氨氯地平片', dose: '5mg', time: '08:00', type: '胶囊', status: 'taken', takenTime: '08:12' },
      { id: 2, name: '阿司匹林肠溶片', dose: '100mg', time: '08:00', type: '片剂', status: 'taken', takenTime: '08:12' },
      { id: 3, name: '阿托伐他汀钙片', dose: '20mg', time: '13:00', type: '片剂', status: 'taken', takenTime: '13:05' },
      { id: 4, name: '氨氯地平片', dose: '5mg', time: '18:00', type: '胶囊', status: 'pending' },
      { id: 5, name: '维生素D', dose: '400IU', time: '18:00', type: '软胶囊', status: 'pending' },
    ],
    weeklyAdherence: 85,
    stats: { taken: 21, missed: 2, late: 1 },
  },
})
